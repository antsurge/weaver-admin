-- =====================================================================
-- 管理员数据权限迁移脚本：admin.data_scope_override → data_permission 规则绑定
-- 数据库：kratos_admin（如有改动请同步脚本中的 USE 语句）
--
-- 背景：
--   v1.x 管理员支持通过 data_scope_override 字段（ALL/DEPT/DEPT_AND_CHILD/SELF，
--   空=跟随角色）覆盖其数据范围；v2.0 起取消该字段，管理员的数据范围改为直接
--   绑定 data_permission 规则（admin_data_permission 关联表），空 = 跟随角色。
--
-- 迁移口径（data_scope_override → 规则 scope_type）：
--   ALL            → '1' 全部数据
--   DEPT           → '3' 本部门数据
--   DEPT_AND_CHILD → '4' 本部门及以下数据
--   SELF           → '5' 仅本人数据
--
-- 说明：
--   若角色迁移脚本（migrate-role-data-scope-to-rules.sql）已先执行过，则映射
--   规则已存在，本脚本会复用同 scope_type 规则，不会重复创建。
--
-- 执行时机【重要】：
--   新版服务 ent 自动迁移开启了 WithDropColumn(true)，服务一启动就会
--   DROP admin.data_scope_override 列。因此本脚本务必在升级服务【之前】执行。
--   新装/从未使用 data_scope_override 的环境：无需迁移，脚本会自动跳过数据部分。
--
-- 幂等性：可重复执行；规则与绑定均按条件去重，不会重复插入。
--
-- 执行方式：mysql -uroot kratos_admin < migrate-admin-data-scope-to-rules.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 0. 环境检测：admin 表 data_scope_override 列、data_permission 表是否存在
-- ---------------------------------------------------------------------
SET @has_ds_col := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'admin'
      AND column_name = 'data_scope_override'
);

SET @has_dp_table := (
    SELECT COUNT(*) FROM information_schema.tables
    WHERE table_schema = DATABASE()
      AND table_name = 'data_permission'
);

SELECT
    IF(@has_ds_col > 0,
       'OK: admin 表存在 data_scope_override 列，将执行数据迁移',
       'SKIP: admin 表已无 data_scope_override 列（或为新装环境），跳过数据迁移')
    AS data_scope_check,
    IF(@has_dp_table > 0,
       'OK: data_permission 表存在',
       'WARNING: data_permission 表不存在，数据迁移将被跳过')
    AS data_permission_check;

-- ---------------------------------------------------------------------
-- 1. 确保管理员-数据权限关联表存在（与 ent 自动迁移生成的 DDL 对齐）
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS admin_data_permission (
    `id` varchar(255) NOT NULL COMMENT '主键ID',
    `tenant_id` varchar(36) NOT NULL DEFAULT 'default' COMMENT '租户ID',
    `created_at` datetime(3) NULL COMMENT '创建时间',
    `admin_id` varchar(255) NOT NULL COMMENT '管理员ID',
    `data_permission_id` varchar(36) NOT NULL COMMENT '数据权限规则ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `admin_data_permission_admin_id_data_permission_id` (`admin_id`, `data_permission_id`),
    KEY `admin_data_permission_tenant_id` (`tenant_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '管理员数据权限规则关联表';

-- ---------------------------------------------------------------------
-- 2.1 为每个 (租户, 数据范围) 补一条【迁移映射规则】（若该租户下尚无同等 scope_type 规则）
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
       CONCAT(''由管理员 data_scope_override 迁移生成（'', t.scope_name, ''）''),
       NOW(6), NOW(6)
FROM (
    SELECT a.tenant_id,
           CASE a.data_scope_override
               WHEN ''ALL'' THEN ''1''
               WHEN ''DEPT'' THEN ''3''
               WHEN ''DEPT_AND_CHILD'' THEN ''4''
               WHEN ''SELF'' THEN ''5''
           END AS scope_type,
           CASE a.data_scope_override
               WHEN ''ALL'' THEN ''全部数据''
               WHEN ''DEPT'' THEN ''本部门数据''
               WHEN ''DEPT_AND_CHILD'' THEN ''本部门及以下数据''
               WHEN ''SELF'' THEN ''仅本人数据''
           END AS scope_name
    FROM admin a
    WHERE a.deleted_at IS NULL
      AND a.data_scope_override IS NOT NULL AND a.data_scope_override != ''''
    GROUP BY a.tenant_id, scope_type, scope_name
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
-- 2.2 按管理员绑定规则：每个管理员的 data_scope_override 对应其租户下匹配 scope_type 的规则
-- ---------------------------------------------------------------------
SET @sql_bind := IF(
    @has_ds_col > 0 AND @has_dp_table > 0,
    '
INSERT INTO admin_data_permission (id, tenant_id, created_at, admin_id, data_permission_id)
SELECT CONCAT(''adp'', REPLACE(UUID(), ''-'', '''')),
       a.tenant_id,
       NOW(6),
       a.id,
       dp.id
FROM admin a
JOIN data_permission dp ON dp.tenant_id = a.tenant_id AND dp.deleted_at IS NULL
     AND dp.scope_type = CASE a.data_scope_override
         WHEN ''ALL'' THEN ''1''
         WHEN ''DEPT'' THEN ''3''
         WHEN ''DEPT_AND_CHILD'' THEN ''4''
         WHEN ''SELF'' THEN ''5''
     END
WHERE a.deleted_at IS NULL
  AND a.data_scope_override IS NOT NULL AND a.data_scope_override != ''''
  AND NOT EXISTS (
      SELECT 1 FROM admin_data_permission b
      WHERE b.admin_id = a.id AND b.data_permission_id = dp.id
  );
',
    'SELECT 1;'
);

PREPARE bind_stmt FROM @sql_bind;
EXECUTE bind_stmt;
DEALLOCATE PREPARE bind_stmt;

-- ---------------------------------------------------------------------
-- 3. 迁移完成校验
-- ---------------------------------------------------------------------
SELECT '管理员-规则绑定统计（按租户）：' AS info;
SELECT b.tenant_id, COUNT(*) AS bind_count
FROM admin_data_permission b
GROUP BY b.tenant_id;

SELECT '迁移完成。新版服务启动后 ent 自动迁移将 DROP admin.data_scope_override 列；' AS tip;