-- 订单菜单初始化 SQL
-- 说明：1) 幂等执行；2) 目录仅当 MenuModule=order 不存在时创建；3) 按钮按生成配置自动挂到菜单下。
-- 角色授权默认给 super_admin，如需其他角色请自行调整。

-- 1. 业务菜单（挂到 MenuModule=order 目录下；若目录不存在先创建）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, weight, status, auth_code, created_at, updated_at)
VALUES (
  'order_menu',
  order_catalog_id,
  '订单',
  'Order',
  '订单',
  '/order/order',
  'lucide:box',
  'menu',
  'order/order/index',
  20,
  'enabled',
  'Order:Order',
  NOW(6), NOW(6)
);

-- 2. 按钮（按生成配置）
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_list', 'order_menu', '列表', 'List', '列表', 'action', 1, 'enabled', 'Order:Order:List', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_detail', 'order_menu', '详情', 'Detail', '详情', 'action', 2, 'enabled', 'Order:Order:Info', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_create', 'order_menu', '创建', 'Create', '创建', 'action', 3, 'enabled', 'Order:Order:Create', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_edit', 'order_menu', '编辑', 'Edit', '编辑', 'action', 4, 'enabled', 'Order:Order:Edit', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_delete', 'order_menu', '删除', 'Delete', '删除', 'action', 5, 'enabled', 'Order:Order:Delete', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_batch_delete', 'order_menu', '批量删除', 'BatchDelete', '批量删除', 'action', 6, 'enabled', 'Order:Order:BatchDelete', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_status', 'order_menu', '状态', 'Status', '状态', 'action', 7, 'enabled', 'Order:Order:SwitchStatus', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_import', 'order_menu', '导入', 'Import', '导入', 'action', 8, 'enabled', 'Order:Order:Import', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('order_btn_export', 'order_menu', '导出', 'Export', '导出', 'action', 9, 'enabled', 'Order:Order:Export', NOW(6), NOW(6));

-- 3. 授权给超管角色（幂等）
INSERT IGNORE INTO role_menu (id, role_id, menu_id, created_at)
SELECT 'order_rm_super', r.id, 'order_menu', NOW(6)
FROM role r WHERE r.is_super_admin = 1
  AND NOT EXISTS (SELECT 1 FROM role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = 'order_menu');
