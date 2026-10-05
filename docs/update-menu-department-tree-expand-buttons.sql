-- =====================================================================
-- 部门管理（Department）»展开全部/折叠全部«按钮节点初始化脚本
-- 数据库：kratos_admin，表：menu
-- 目的：为"部门管理"菜单节点补充两个按钮级权限（type='action'）：
--         * 展开全部  auth_code = 'Organization:Department:ExpandAll'
--         * 折叠全部  auth_code = 'Organization:Department:CollapseAll'
--       使前端 v-access:code 指令（PermissionAuthCode.Department.ExpandAll /
--       CollapseAll）在部门管理页面对应工具栏按钮生效。
-- 说明：
--   * 展开/折叠为纯前端树形视图操作，无后端接口，因此本脚本不绑定
--     menu_api_permission（不需要 api_interface）。
--   * 幂等：可重复执行；已存在的按钮不会被重复插入。
--   * 本脚本针对 default 租户（tenant_id='default'）生效；
--     若使用 multi 租户，请把 tenant_id 换成目标租户或去掉条件。
--   * 定位"部门管理"菜单节点：优先 auth_code='Organization:Department'，
--     其次 name='部门管理' / title='organization.department.title' /
--     component='organization/department/index'。
--   * 若找不到"部门管理"菜单节点，脚本会给出 WARNING 且不插入任何按钮，
--     请先通过「菜单管理」创建该菜单节点后再执行。
--   * 按钮插入后，还需在「角色管理」中为对应角色勾选这两个按钮权限，
--     或者通过菜单管理确认按钮已挂到部门管理节点下。
-- 执行方式：mysql -uroot kratos_admin < update-menu-department-tree-expand-buttons.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 0. 定位"部门管理"菜单节点
-- ---------------------------------------------------------------------
SET @department_menu_id := (
    SELECT id FROM menu
    WHERE tenant_id = 'default'
      AND (
            auth_code = 'Organization:Department'
         OR name = '部门管理'
         OR title = 'organization.department.title'
         OR component = 'organization/department/index'
      )
    ORDER BY (auth_code = 'Organization:Department') DESC
    LIMIT 1
);

SELECT IF(@department_menu_id IS NOT NULL,
          CONCAT('OK: 已定位部门管理菜单节点 id=', @department_menu_id),
          'WARNING: 未找到部门管理菜单节点，请先在「菜单管理」中创建后再执行本脚本')
       AS department_menu_locate;

-- ---------------------------------------------------------------------
-- 1. 按钮节点（父菜单不存在时不插入，避免孤儿数据）
--    每个按钮在数据库中唯一标识：tenant_id='default' AND parent_id=部门菜单 AND auth_code
-- ---------------------------------------------------------------------

-- 1.1 按钮：展开全部（ExpandAll）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('dept', REPLACE(UUID(), '-', '')), @department_menu_id, '展开全部', '', '展开全部', '', '', 'action', '', 'Organization:Department:ExpandAll', 80, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @department_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @department_menu_id AND auth_code = 'Organization:Department:ExpandAll');

-- 1.2 按钮：折叠全部（CollapseAll）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, auth_code, weight, status, tenant_id, created_at, updated_at)
SELECT CONCAT('dept', REPLACE(UUID(), '-', '')), @department_menu_id, '折叠全部', '', '折叠全部', '', '', 'action', '', 'Organization:Department:CollapseAll', 90, 'enabled', 'default', NOW(6), NOW(6)
FROM DUAL
WHERE @department_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM menu WHERE tenant_id = 'default' AND parent_id = @department_menu_id AND auth_code = 'Organization:Department:CollapseAll');

-- ---------------------------------------------------------------------
-- 2. 验证：部门管理菜单下的按钮节点
-- ---------------------------------------------------------------------
SELECT
    m.id        AS menu_id,
    m.name      AS menu_name,
    m.auth_code AS auth_code,
    m.weight    AS weight,
    m.status    AS status
FROM menu m
WHERE m.tenant_id = 'default'
  AND m.parent_id = @department_menu_id
  AND m.auth_code IN ('Organization:Department:ExpandAll', 'Organization:Department:CollapseAll')
ORDER BY m.weight;