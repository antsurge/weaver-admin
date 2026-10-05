-- 测试2菜单初始化 SQL
-- 说明：1) 幂等执行；2) 目录仅当 MenuModule=system 不存在时创建；3) 按钮按生成配置自动挂到菜单下。
-- 角色授权默认给 super_admin，如需其他角色请自行调整。

-- 1. 业务菜单（挂到 MenuModule=system 目录下；若目录不存在先创建）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, weight, status, auth_code, created_at, updated_at)
VALUES (
  'test_menu',
  system_catalog_id,
  '测试2',
  'test',
  '测试2',
  '/system/test',
  'lucide:box',
  'menu',
  'system/test/index',
  20,
  'enabled',
  'System:test',
  NOW(6), NOW(6)
);

-- 2. 按钮（按生成配置）
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_list', 'test_menu', '列表', 'List', '列表', 'action', 1, 'enabled', 'System:test:List', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_detail', 'test_menu', '详情', 'Detail', '详情', 'action', 2, 'enabled', 'System:test:Info', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_create', 'test_menu', '创建', 'Create', '创建', 'action', 3, 'enabled', 'System:test:Create', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_edit', 'test_menu', '编辑', 'Edit', '编辑', 'action', 4, 'enabled', 'System:test:Edit', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_delete', 'test_menu', '删除', 'Delete', '删除', 'action', 5, 'enabled', 'System:test:Delete', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_batch_delete', 'test_menu', '批量删除', 'BatchDelete', '批量删除', 'action', 6, 'enabled', 'System:test:BatchDelete', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_status', 'test_menu', '状态', 'Status', '状态', 'action', 7, 'enabled', 'System:test:SwitchStatus', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_import', 'test_menu', '导入', 'Import', '导入', 'action', 8, 'enabled', 'System:test:Import', NOW(6), NOW(6));
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('test_btn_export', 'test_menu', '导出', 'Export', '导出', 'action', 9, 'enabled', 'System:test:Export', NOW(6), NOW(6));

-- 3. 授权给超管角色（幂等）
INSERT IGNORE INTO role_menu (id, role_id, menu_id, created_at)
SELECT 'test_rm_super', r.id, 'test_menu', NOW(6)
FROM role r WHERE r.is_super_admin = 1
  AND NOT EXISTS (SELECT 1 FROM role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = 'test_menu');
