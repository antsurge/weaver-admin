-- =====================================================================
-- 安全中心（Security）菜单迁移脚本
-- 数据库：kratos_admin，表：menu / role_menus
-- 目的：把原挂在"系统管理"下的在线用户/安全策略
--       重挂到新顶级目录"安全中心"（path=/security）下。
-- 注：操作日志/登录日志已拆分到"日志中心"，请另行执行 update-menu-logs.sql。
-- 说明：
--   * 幂等：可重复执行；已迁移过的记录不会被重复挂载。
--   * 本脚本针对 default 租户（tenant_id='default'）生效；
--     若使用 multi 租户，请把 tenant_id 换成目标租户或去掉条件。
-- 执行方式：mysql -uroot kratos_admin < update-menu-security.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 1. 确保"安全中心"目录存在（不存在才插入）
--    type='catalog', code='Security', path='/security'
-- ---------------------------------------------------------------------
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, weight, status, tenant_id, created_at, updated_at)
SELECT
    CONCAT('s', REPLACE(UUID(), '-', '')),
    '',
    '安全中心',
    'Security',
    'page.security.title',
    '/security',
    'lucide:shield-check',
    'catalog',
    '',
    6,
    'enabled',
    'default',
    NOW(6),
    NOW(6)
WHERE NOT EXISTS (
    SELECT 1 FROM menu WHERE tenant_id = 'default' AND code = 'Security'
);

-- ---------------------------------------------------------------------
-- 2. 重挂子菜单到安全中心（按 code 定位，name 兜底），更新 path/component
--    幂等：只有 parent_id 仍指向旧目录时才更新
-- ---------------------------------------------------------------------

-- 2.1 在线用户
UPDATE menu
SET parent_id = (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Security') t),
    path = '/security/online',
    component = 'security/online/index',
    title = 'page.system.online'
WHERE tenant_id = 'default'
  AND (code = 'Online' OR name = '在线用户')
  AND parent_id <> (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Security') t);

-- 2.2 安全策略
UPDATE menu
SET parent_id = (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Security') t),
    path = '/security/settings',
    component = 'security/settings/index',
    title = 'page.security.settings'
WHERE tenant_id = 'default'
  AND (code = 'SecuritySettings' OR name = '安全策略')
  AND parent_id <> (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Security') t);

-- ---------------------------------------------------------------------
-- 3. 验证：输出安全中心目录及其子菜单
-- ---------------------------------------------------------------------
SELECT
    child.id,
    child.name,
    child.code,
    child.title,
    child.path,
    child.component
FROM menu child
JOIN menu parent ON parent.id = child.parent_id
WHERE parent.code = 'Security' AND child.tenant_id = 'default'
ORDER BY child.weight;

-- 剩余仍挂在系统管理下的安全类菜单（期望为空，若有说明 name/code 不匹配，可手动核对）
SELECT id, name, code, parent_id, path, component
FROM menu
WHERE tenant_id = 'default'
  AND (name IN ('在线用户', '安全策略')
       OR component LIKE 'security/%')
  AND parent_id <> (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Security') t);
