-- =====================================================================
-- 角色管理（Role）菜单按钮节点初始化脚本
-- 数据库：kratos_admin，表：menu / menu_api_permission / api_interface
-- 目的：为"角色管理"菜单节点补充按钮级权限（type='action'），
--       并将按钮与后端接口（api_interface.code = service|METHOD|path）绑定，
--       使前端 v-access:code 指令（Permission:Role:*）可与后端接口鉴权联动。
-- 说明：
--   * 绑定基于 menu_api_permission（menu_id + api_code）关联，
--     接口信息实时反查 api_interface，接口重导入（全量清空重建）后绑定依然有效。
--   * 幂等：可重复执行；已存在的按钮 / 绑定不会被重复插入。
--   * 本脚本针对 default 租户（tenant_id='default'）生效；
--     若使用 multi 租户，请把 tenant_id 换成目标租户或去掉条件。
--   * 定位"角色管理"菜单节点：优先 auth_code='Permission:Role'，
--     其次 name='角色管理' / title='permission.role.title' / component='permission/role/index'。
--   * 若找不到"角色管理"菜单节点，脚本会给出 WARNING 且不插入任何按钮，
--     请先通过「菜单管理」创建该菜单节点后再执行。
--   * 若 api_interface 中缺少对应接口（未导入 openapi.yaml），绑定会被跳过，
--     请先通过「接口管理」上传 openapi.yaml 后再执行。
-- 执行方式：mysql -uroot kratos_admin < update-menu-role-buttons.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 0. 定位"角色管理"菜单节点
-- ---------------------------------------------------------------------
SET @role_menu_id := (
    SELECT id FROM menu
    WHERE tenant_id = 'default'
      AND (
            auth_code = 'Permission:Role'
         OR name = '角色管理'
         OR title = 'permission.role.title'
         OR component = 'permission/role/index'
      )
    ORDER BY (auth_code = 'Permission:Role') DESC
    LIMIT 1
);

SELECT IF(@role_menu_id IS NOT NULL,
          CONCAT('OK: 已定位角色管理菜单节点 id=', @role_menu_id),
          'WARNING: 未找到角色管理菜单节点，请先在「菜单管理」中创建后再执行本脚本')
       AS role_menu_locate;

-- ---------------------------------------------------------------------
-- 1. 按钮节点（父菜单不存在时不插入，避免孤儿数据）
--    每个按钮在数据库中唯一标识：tenant_id='default' AND parent_id=角色菜单 AND auth_code
-- ---------------------------------------------------------------------

-- 1.1 按钮：角色列表（List）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('r', REPLACE(UUID(), '-', '')), @role_menu_id, '角色列表', '', '角色列表', '', '', 'action', '', 'Permission:Role:List', 10, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @role_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @role_menu_id AND auth_code = 'Permission:Role:List');

-- 1.2 按钮：查看详情（Info）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('r', REPLACE(UUID(), '-', '')), @role_menu_id, '角色详情', '', '角色详情', '', '', 'action', '', 'Permission:Role:Info', 20, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @role_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @role_menu_id AND auth_code = 'Permission:Role:Info');

-- 1.3 按钮：新增（Create）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('r', REPLACE(UUID(), '-', '')), @role_menu_id, '新增角色', '', '新增', '', '', 'action', '', 'Permission:Role:Create', 30, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @role_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @role_menu_id AND auth_code = 'Permission:Role:Create');

-- 1.4 按钮：编辑（Edit）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('r', REPLACE(UUID(), '-', '')), @role_menu_id, '编辑角色', '', '编辑', '', '', 'action', '', 'Permission:Role:Edit', 40, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @role_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @role_menu_id AND auth_code = 'Permission:Role:Edit');

-- 1.5 按钮：删除（Delete）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('r', REPLACE(UUID(), '-', '')), @role_menu_id, '删除角色', '', '删除', '', '', 'action', '', 'Permission:Role:Delete', 50, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @role_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @role_menu_id AND auth_code = 'Permission:Role:Delete');

-- 1.6 按钮：批量删除（BatchDelete）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('r', REPLACE(UUID(), '-', '')), @role_menu_id, '批量删除角色', '', '批量删除', '', '', 'action', '', 'Permission:Role:BatchDelete', 60, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @role_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @role_menu_id AND auth_code = 'Permission:Role:BatchDelete');

-- 1.7 按钮：启用/禁用（SwitchStatus）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('r', REPLACE(UUID(), '-', '')), @role_menu_id, '启用/禁用角色', '', '启用/禁用', '', '', 'action', '', 'Permission:Role:SwitchStatus', 70, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @role_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @role_menu_id AND auth_code = 'Permission:Role:SwitchStatus');

-- ---------------------------------------------------------------------
-- 2. 按钮 ↔ 接口绑定（menu_api_permission 表，幂等）
--    接口以 api_interface.code = service|METHOD|path 关联（不依赖接口 id，
--    接口重导入后 id 变化但 code 稳定，绑定不失效）。
--    只绑定 api_interface 中已存在的接口（需先通过「接口管理」导入 openapi.yaml）。
-- ---------------------------------------------------------------------

-- 2.1 List -> GET /admin/v1/role
INSERT INTO menu_api_permission (id, menu_id, api_code, tenant_id, created_at)
SELECT CONCAT('b', REPLACE(UUID(), '-', '')), m.id, i.code, m.tenant_id, NOW(6)
FROM menu m
JOIN api_interface i ON i.code = 'PermissionService|GET|/admin/v1/role'
WHERE m.tenant_id = 'default' AND m.auth_code = 'Permission:Role:List'
  AND NOT EXISTS (
      SELECT 1 FROM menu_api_permission b
      WHERE b.menu_id = m.id AND b.api_code = i.code
  );

-- 2.2 Info -> GET /admin/v1/role/{id}
INSERT INTO menu_api_permission (id, menu_id, api_code, tenant_id, created_at)
SELECT CONCAT('b', REPLACE(UUID(), '-', '')), m.id, i.code, m.tenant_id, NOW(6)
FROM menu m
JOIN api_interface i ON i.code = 'PermissionService|GET|/admin/v1/role/{id}'
WHERE m.tenant_id = 'default' AND m.auth_code = 'Permission:Role:Info'
  AND NOT EXISTS (
      SELECT 1 FROM menu_api_permission b
      WHERE b.menu_id = m.id AND b.api_code = i.code
  );

-- 2.3 Create -> POST /admin/v1/role
INSERT INTO menu_api_permission (id, menu_id, api_code, tenant_id, created_at)
SELECT CONCAT('b', REPLACE(UUID(), '-', '')), m.id, i.code, m.tenant_id, NOW(6)
FROM menu m
JOIN api_interface i ON i.code = 'PermissionService|POST|/admin/v1/role'
WHERE m.tenant_id = 'default' AND m.auth_code = 'Permission:Role:Create'
  AND NOT EXISTS (
      SELECT 1 FROM menu_api_permission b
      WHERE b.menu_id = m.id AND b.api_code = i.code
  );

-- 2.4 Edit -> PUT /admin/v1/role/{id}
INSERT INTO menu_api_permission (id, menu_id, api_code, tenant_id, created_at)
SELECT CONCAT('b', REPLACE(UUID(), '-', '')), m.id, i.code, m.tenant_id, NOW(6)
FROM menu m
JOIN api_interface i ON i.code = 'PermissionService|PUT|/admin/v1/role/{id}'
WHERE m.tenant_id = 'default' AND m.auth_code = 'Permission:Role:Edit'
  AND NOT EXISTS (
      SELECT 1 FROM menu_api_permission b
      WHERE b.menu_id = m.id AND b.api_code = i.code
  );

-- 2.5 Delete -> DELETE /admin/v1/role（删除 & 批量删除共用）
INSERT INTO menu_api_permission (id, menu_id, api_code, tenant_id, created_at)
SELECT CONCAT('b', REPLACE(UUID(), '-', '')), m.id, i.code, m.tenant_id, NOW(6)
FROM menu m
JOIN api_interface i ON i.code = 'PermissionService|DELETE|/admin/v1/role'
WHERE m.tenant_id = 'default' AND m.auth_code IN ('Permission:Role:Delete', 'Permission:Role:BatchDelete')
  AND NOT EXISTS (
      SELECT 1 FROM menu_api_permission b
      WHERE b.menu_id = m.id AND b.api_code = i.code
  );

-- 2.6 SwitchStatus -> PUT /admin/v1/role/{id}/status
INSERT INTO menu_api_permission (id, menu_id, api_code, tenant_id, created_at)
SELECT CONCAT('b', REPLACE(UUID(), '-', '')), m.id, i.code, m.tenant_id, NOW(6)
FROM menu m
JOIN api_interface i ON i.code = 'PermissionService|PUT|/admin/v1/role/{id}/status'
WHERE m.tenant_id = 'default' AND m.auth_code = 'Permission:Role:SwitchStatus'
  AND NOT EXISTS (
      SELECT 1 FROM menu_api_permission b
      WHERE b.menu_id = m.id AND b.api_code = i.code
  );

-- ---------------------------------------------------------------------
-- 3. 验证：角色管理菜单下的按钮节点及其绑定的接口
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
  AND m.parent_id = @role_menu_id
ORDER BY m.weight, i.path;
