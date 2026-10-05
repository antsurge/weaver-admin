-- =====================================================================
-- 菜单按钮绑定迁移脚本：旧快照表 → 新绑定表
-- 数据库：kratos_admin
-- 旧结构：api_permission（快照）+ menu_api_permissions（menu_id ↔ api_permission_id）
-- 新结构：menu_api_permission（menu_id + api_code，接口信息实时反查 api_interface）
--
-- 迁移说明：
--   * 旧 api_permission 快照表包含 code（service|METHOD|path）字段，
--     迁移时取出 code 写入新绑定表 menu_api_permission。
--   * 新表 id 用 XID 风格前缀（36 位内），与代码中 uuid.GenerateXID() 生成的 ID 共存无冲突。
--   * 幂等：已存在的绑定（menu_id + api_code）不会重复插入。
--   * 建议在服务升级并完成自动迁移（新增 menu_api_permission 表）后执行本脚本。
--   * 迁移完成后旧表 api_permission / menu_api_permissions 可手动 DROP，
--     服务端代码已不再读写这两张表。
-- 执行方式：mysql -uroot kratos_admin < migrate-menu-api-code.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 1. 迁移绑定数据：menu_api_permissions（旧关联表）→ menu_api_permission（新绑定表）
--    通过 api_permission.code 把旧的 api_permission_id 转成新的 api_code
-- ---------------------------------------------------------------------
INSERT INTO menu_api_permission (id, menu_id, api_code, tenant_id, created_at)
SELECT
    CONCAT('b', REPLACE(UUID(), '-', '')) AS id,
    mmp.menu_id                           AS menu_id,
    ap.code                               AS api_code,
    m.tenant_id                           AS tenant_id,
    COALESCE(mmp.created_at, NOW(6))      AS created_at
FROM menu_api_permissions mmp
JOIN api_permission ap ON ap.id = mmp.api_permission_id
JOIN menu m ON m.id = mmp.menu_id
WHERE ap.code IS NOT NULL AND ap.code != ''
  AND NOT EXISTS (
      SELECT 1 FROM menu_api_permission b
      WHERE b.menu_id = mmp.menu_id AND b.api_code = ap.code
  );

-- ---------------------------------------------------------------------
-- 2. 验证：迁移后的绑定数量（应等于旧关联表中 code 非空的数量）
-- ---------------------------------------------------------------------
SELECT
    (SELECT COUNT(*) FROM menu_api_permission) AS new_bindings_total,
    (SELECT COUNT(*) FROM menu_api_permissions
        JOIN api_permission ap ON ap.id = menu_api_permissions.api_permission_id
        WHERE ap.code IS NOT NULL AND ap.code != '') AS old_bindings_with_code;

-- 若确认无误，可手动清理旧表：
--   DROP TABLE IF EXISTS menu_api_permissions;
--   DROP TABLE IF EXISTS api_permission;
