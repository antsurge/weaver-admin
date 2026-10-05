-- =====================================================================
-- 角色数据权限迁移脚本：role.data_scope → data_permission 规则绑定
-- 数据库：kratos_admin（如有改动请同步脚本中的 USE 语句）
--
-- 背景：
--   v1.x 角色通过四值字段 data_scope（ALL / DEPT / DEPT_AND_CHILD / SELF）
--   直接控制数据范围；v2.0 起取消该字段，角色的数据范围完全由所绑定的
--   data_permission 规则（role_data_permission 关联表）决定。
--
-- 迁移口径（data_scope → 规则 scope_type）：
--   ALL            → '1' 全部数据
--   DEPT           → '3' 本部门数据
--   DEPT_AND_CHILD → '4' 本部门及以下数据
--   SELF           → '5' 仅本人数据
--
-- 执行时机【重要】：
--   新版服务 ent 自动迁移开启了 WithDropColumn(true)，服务一启动就会
--   DROP role.data_scope 列。因此本脚本务必在升级服务【之前】执行，
--   否则旧值将被删除、无法迁移。
--   新装/从未使用 data_scope 的环境：无需迁移，脚本会自动跳过数据部分。
--
-- 幂等性：可重复执行；规则与绑定均按条件去重，不会重复插入。
--
-- 执行方式：mysql -uroot kratos_admin < migrate-role-data-scope-to-rules.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 0. 环境检测：role_data_permission / data_permission 表与 data_scope 列
-- ---------------------------------------------------------------------
-- role 表是否仍存在 data_scope 列（旧库升级场景 > 0）
SET @has_ds_col := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'role'
      AND column_name = 'data_scope'
);

-- data_permission 表是否存在（旧库若已使用数据权限模块则存在）
SET @has_dp_table := (
    SELECT COUNT(*) FROM information_schema.tables
    WHERE table_schema = DATABASE()
      AND table_name = 'data_permission'
);

SELECT
    IF(@has_ds_col > 0,
       CONCAT('OK: role 表存在 data_scope 列，将执行数据迁移'),
       'SKIP: role 表已无 data_scope 列（或为新装环境），跳过数据迁移')
    AS data_scope_check,
    IF(@has_dp_table > 0,
       'OK: data_permission 表存在',
       'WARNING: data_permission 表不存在，数据迁移将被跳过')
    AS data_permission_check;

-- ---------------------------------------------------------------------
-- 1. 确保角色-数据权限关联表存在（与 ent 自动迁移生成的 DDL 对齐，
--    避免在服务启动前手动建表时缺失）
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS role_data_permission (
    `id` varchar(255) NOT NULL COMMENT '主键ID',
    `tenant_id` varchar(36) NOT NULL DEFAULT 'default' COMMENT '租户ID',
    `created_at` datetime(3) NULL COMMENT '创建时间',
    `role_id` varchar(255) NOT NULL COMMENT '角色ID',
    `data_permission_id` varchar(36) NOT NULL COMMENT '数据权限规则ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `role_data_permission_role_id_data_permission_id` (`role_id`, `data_permission_id`),
    KEY `role_data_permission_tenant_id` (`tenant_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '角色数据权限规则关联表';

-- ---------------------------------------------------------------------
-- 2.1 为每个 (租户, 数据范围) 补一条【迁移映射规则】
--     仅当该租户下尚无同等 scope_type 的规则时插入（幂等 + code 用 UUID 片段避免唯一冲突）
-- ---------------------------------------------------------------------
SET @sql_mk_rule := IF(
    @has_ds_col > 0 AND @has_dp_table > 0,
    '
INSERT INTO data_permission (id, tenant_id, name, code, scope_type, dept_ids, role_ids, status, remark, created_at, updated_at)
SELECT CONCAT(''md'', REPLACE(UUID(), ''-'', '''')),
       t.tenant_id,
       t.scope_name,
       CONCAT(''MIGRATED_SCOPE_'', t.scope_type, ''_'', LEFT(REPLACE(UUID(), ''-'', ''''), 4)),
       t.scope_type,
       ''[]'', ''[]'', ''enabled'',
       CONCAT(''由角色 data_scope 迁移生成（'', t.scope_name, ''）''),
       NOW(6), NOW(6)
FROM (
    SELECT r.tenant_id,
           CASE r.data_scope
               WHEN ''ALL'' THEN ''1''
               WHEN ''DEPT'' THEN ''3''
               WHEN ''DEPT_AND_CHILD'' THEN ''4''
               WHEN ''SELF'' THEN ''5''
           END AS scope_type,
           CASE r.data_scope
               WHEN ''ALL'' THEN ''全部数据''
               WHEN ''DEPT'' THEN ''本部门数据''
               WHEN ''DEPT_AND_CHILD'' THEN ''本部门及以下数据''
               WHEN ''SELF'' THEN ''仅本人数据''
           END AS scope_name
    FROM role r
    WHERE r.deleted_at IS NULL AND r.data_scope IS NOT NULL AND r.data_scope != ''''
    GROUP BY r.tenant_id, scope_type, scope_name
) t
WHERE NOT EXISTS (
    SELECT 1 FROM data_permission dp
    WHERE dp.tenant_id = t.tenant_id AND dp.scope_type = t.scope_type AND dp.deleted_at IS NULL
);
',
    'SELECT 1;'
);

PREPARE mk_rule FROM @sql_mk_rule;
EXECUTE mk_rule;
DEALLOCATE PREPARE mk_rule;

-- ---------------------------------------------------------------------
-- 2.2 按角色绑定规则：每个角色的 data_scope 对应其租户下匹配 scope_type 的规则
-- ---------------------------------------------------------------------
SET @sql_bind := IF(
    @has_ds_col > 0 AND @has_dp_table > 0,
    '
INSERT INTO role_data_permission (id, tenant_id, created_at, role_id, data_permission_id)
SELECT CONCAT(''rdp'', REPLACE(UUID(), ''-'', '''')),
       r.tenant_id,
       NOW(6),
       r.id,
       dp.id
FROM role r
JOIN data_permission dp ON dp.tenant_id = r.tenant_id AND dp.deleted_at IS NULL
     AND dp.scope_type = CASE r.data_scope
         WHEN ''ALL'' THEN ''1''
         WHEN ''DEPT'' THEN ''3''
         WHEN ''DEPT_AND_CHILD'' THEN ''4''
         WHEN ''SELF'' THEN ''5''
     END
WHERE r.deleted_at IS NULL AND r.data_scope IS NOT NULL AND r.data_scope != ''''
  AND NOT EXISTS (
      SELECT 1 FROM role_data_permission b
      WHERE b.role_id = r.id AND b.data_permission_id = dp.id
  );
',
    'SELECT 1;'
);

PREPARE bind_stmt FROM @sql_bind;
EXECUTE bind_stmt;
DEALLOCATE PREPARE bind_stmt;

-- ---------------------------------------------------------------------
-- 3. 迁移完成校验：展示生成的映射规则与绑定数量
-- ---------------------------------------------------------------------
SELECT '以下为本次迁移生成的映射规则：' AS info;
SELECT id, tenant_id, name, code, scope_type FROM data_permission
WHERE code LIKE 'MIGRATED_SCOPE_%' ORDER BY tenant_id, scope_type;

SELECT '角色-规则绑定统计（按租户）：' AS info;
SELECT b.tenant_id, COUNT(*) AS bind_count
FROM role_data_permission b
GROUP BY b.tenant_id;

SELECT '迁移完成。新版服务启动后 ent 自动迁移将 DROP role.data_scope 列；' AS tip;