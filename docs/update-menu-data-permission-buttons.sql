-- =====================================================================
-- 数据权限（DataPermission）菜单按钮节点初始化脚本
-- 数据库：kratos_admin，表：menu / menu_api_permission / api_interface
-- 目的：为"数据权限"菜单节点补充「启用/禁用」按钮级权限（type='action'，
--       auth_code='Permission:DataPermission:SwitchStatus'），
--       并将按钮与后端接口（api_interface.code = service|METHOD|path）绑定，
--       使前端 v-access:code 指令（Permission:DataPermission:SwitchStatus）
--       可与后端接口鉴权联动。
-- 说明：
--   * 绑定基于 menu_api_permission（menu_id + api_code）关联，
--     接口信息实时反查 api_interface，接口重导入（全量清空重建）后绑定依然有效。
--   * 幂等：可重复执行；已存在的按钮 / 绑定不会被重复插入。
--   * 本脚本针对 default 租户（tenant_id='default'）生效；
--     若使用 multi 租户，请把 tenant_id 换成目标租户或去掉条件。
--   * 定位"数据权限"菜单节点：优先 auth_code='Permission:DataPermission'，
--     其次 name='数据权限' / component='permission/data-permission/index'。
--   * 若找不到"数据权限"菜单节点，脚本会给出 WARNING 且不插入任何按钮，
--     请先通过「菜单管理」创建该菜单节点后再执行。
--   * 若 api_interface 中缺少对应接口（未导入 openapi.yaml），绑定会被跳过，
--     请先通过「接口管理」上传 openapi.yaml 后再执行。
-- 执行方式：mysql -uroot kratos_admin < update-menu-data-permission-buttons.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 0. 定位"数据权限"菜单节点
-- ---------------------------------------------------------------------
SET @dp_menu_id := (
    SELECT id FROM menu
    WHERE tenant_id = 'default'
      AND (
            auth_code = 'Permission:DataPermission'
         OR name = '数据权限'
         OR component = 'permission/data-permission/index'
      )
    ORDER BY (auth_code = 'Permission:DataPermission') DESC
    LIMIT 1
);

SELECT IF(@dp_menu_id IS NOT NULL,
          CONCAT('OK: 已定位数据权限菜单节点 id=', @dp_menu_id),
          'WARNING: 未找到数据权限菜单节点，请先在「菜单管理」中创建后再执行本脚本')
       AS dp_menu_locate;

-- ---------------------------------------------------------------------
-- 1. 按钮节点（父菜单不存在时不插入，避免孤儿数据）
--    每个按钮在数据库中唯一标识：tenant_id='default' AND parent_id=数据权限菜单 AND auth_code
-- ---------------------------------------------------------------------

-- 1.1 按钮：启用/禁用（SwitchStatus）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('d', REPLACE(UUID(), '-', '')), @dp_menu_id, '启用/禁用数据权限', '', '启用/禁用', '', '', 'action', '', 'Permission:DataPermission:SwitchStatus', 70, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @dp_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @dp_menu_id AND auth_code = 'Permission:DataPermission:SwitchStatus');

-- ---------------------------------------------------------------------
-- 2. 按钮 ↔ 接口绑定（menu_api_permission 表，幂等）
--    接口以 api_interface.code = service|METHOD|path 关联（不依赖接口 id，
--    接口重导入后 id 变化但 code 稳定，绑定不失效）。
--    只绑定 api_interface 中已存在的接口（需先通过「接口管理」导入 openapi.yaml）。
-- ---------------------------------------------------------------------

-- 2.1 SwitchStatus -> PUT /admin/v1/data-permission/{id}/status
INSERT INTO menu_api_permission (id, menu_id, api_code, tenant_id, created_at)
SELECT CONCAT('b', REPLACE(UUID(), '-', '')), m.id, i.code, m.tenant_id, NOW(6)
FROM menu m
JOIN api_interface i ON i.code = 'PermissionService|PUT|/admin/v1/data-permission/{id}/status'
WHERE m.tenant_id = 'default' AND m.auth_code = 'Permission:DataPermission:SwitchStatus'
  AND NOT EXISTS (
      SELECT 1 FROM menu_api_permission b
      WHERE b.menu_id = m.id AND b.api_code = i.code
  );

-- ---------------------------------------------------------------------
-- 3. 验证：数据权限菜单下的按钮节点及其绑定的接口
-- ---------------------------------------------------------------------
SELECT
    m.id        AS menu_id,
    m.name      AS menu_name,
    m.auth_code AS auth_code,
    i.method    AS api_method,
    i.path      AS api_path,
    i.summary   AS api_summary
FROM menu m
LEFT JOIN menu_api_permission b ON b.menu_id = m.id
LEFT JOIN api_interface i ON i.code = b.api_code
WHERE m.tenant_id = 'default'
  AND m.parent_id = @dp_menu_id
ORDER BY m.weight, i.path;
