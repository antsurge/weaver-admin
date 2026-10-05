-- =====================================================================
-- 日志中心（Logs）菜单迁移脚本
-- 数据库：kratos_admin，表：menu / role_menus
-- 目的：
--   1. 把原挂在"安全中心"下的操作日志/登录日志
--      重挂到新顶级目录"日志中心"（path=/logs）下。
--   2. 把原挂在"系统管理"下的任务日志一并迁入日志中心，
--      并修正其不规范记录（name='JobLog'、code=''、title='system.job_log.title'）。
-- 说明：
--   * 幂等：可重复执行；已迁移过的记录不会被重复挂载。
--   * 本脚本针对 default 租户（tenant_id='default'）生效；
--     若使用 multi 租户，请把 tenant_id 换成目标租户或去掉条件。
-- 执行方式：mysql -uroot kratos_admin < update-menu-logs.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 1. 确保"日志中心"目录存在（不存在才插入）
--    type='catalog', code='Log', path='/logs'
-- ---------------------------------------------------------------------
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, weight, status, tenant_id, created_at, updated_at)
SELECT
    CONCAT('l', REPLACE(UUID(), '-', '')),
    '',
    '日志中心',
    'Log',
    'page.log.title',
    '/logs',
    'lucide:scroll-text',
    'catalog',
    '',
    8,
    'enabled',
    'default',
    NOW(6),
    NOW(6)
WHERE NOT EXISTS (
    SELECT 1 FROM menu WHERE tenant_id = 'default' AND code = 'Log'
);

-- ---------------------------------------------------------------------
-- 2. 重挂子菜单到日志中心（按 code 定位，name/path/component 兜底），更新 path/component
--    幂等：只有 parent_id 仍指向旧目录（安全中心/系统管理）时才更新
-- ---------------------------------------------------------------------

-- 2.1 操作日志
UPDATE menu
SET parent_id = (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Log') t),
    path = '/logs/operation-log',
    component = 'log/operation-log/index',
    title = 'page.system.operationLog'
WHERE tenant_id = 'default'
  AND (code = 'OperationLog' OR name = '操作日志')
  AND parent_id <> (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Log') t);

-- 2.2 登录日志
UPDATE menu
SET parent_id = (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Log') t),
    path = '/logs/login-log',
    component = 'log/login-log/index',
    title = 'page.system.loginLog'
WHERE tenant_id = 'default'
  AND (code = 'LoginLog' OR name = '登录日志')
  AND parent_id <> (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Log') t);

-- 2.3 任务日志
-- 历史记录可能不规范（name='JobLog'、code=''、title='system.job_log.title'），
-- 按 code/name/path/component 多条件兜底定位，一并修正为规范值并重挂。
UPDATE menu
SET parent_id = (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Log') t),
    name = '任务日志',
    code = 'JobLog',
    title = 'page.system.jobLog',
    path = '/logs/job-log',
    component = 'log/job-log/index',
    weight = 3
WHERE tenant_id = 'default'
  AND (code = 'JobLog' OR name = 'JobLog' OR name = '任务日志'
       OR path = '/system/job-log' OR component = 'ops/job-log/index' OR component = 'system/job-log/index')
  AND parent_id <> (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Log') t);

-- ---------------------------------------------------------------------
-- 3. 清理历史重复记录（仅限日志菜单）：
--    旧版"系统管理"目录（name='System', code='', weight=0）下残留的英文名
--    OperationLog/LoginLog（code='', component=security/log/*）为废弃记录，
--    但 role_menu 权限可能仍指向它们。先把引用迁移到日志中心下的规范记录，
--    再删除废弃记录。
-- ---------------------------------------------------------------------

-- 3.1 操作日志：把 role_menu 中指向旧记录（security/log/operation-log）的引用改到新记录
UPDATE role_menu rm
JOIN menu old_m ON old_m.id = rm.menu_id
    AND old_m.tenant_id = 'default'
    AND old_m.name = 'OperationLog'
    AND old_m.code = ''
    AND old_m.component LIKE 'security/log/%'
JOIN menu new_m ON new_m.tenant_id = 'default'
    AND new_m.code = 'OperationLog'
    AND new_m.component = 'log/operation-log/index'
SET rm.menu_id = new_m.id;

-- 3.2 登录日志：同上
UPDATE role_menu rm
JOIN menu old_m ON old_m.id = rm.menu_id
    AND old_m.tenant_id = 'default'
    AND old_m.name = 'LoginLog'
    AND old_m.code = ''
    AND old_m.component LIKE 'security/log/%'
JOIN menu new_m ON new_m.tenant_id = 'default'
    AND new_m.code = 'LoginLog'
    AND new_m.component = 'log/login-log/index'
SET rm.menu_id = new_m.id;

-- 3.3 删除废弃的英文名日志菜单记录（避免 UI 重复展示）
DELETE m
FROM menu m
WHERE m.tenant_id = 'default'
  AND m.code = ''
  AND m.name IN ('OperationLog', 'LoginLog')
  AND m.component LIKE 'security/log/%';

-- ---------------------------------------------------------------------
-- 4. 验证：输出日志中心目录及其子菜单
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
WHERE parent.code = 'Log' AND child.tenant_id = 'default'
ORDER BY child.weight;

-- 剩余仍挂在安全中心/系统管理下的日志菜单（期望为空，若有说明 name/code 不匹配，可手动核对）
SELECT id, name, code, parent_id, path, component
FROM menu
WHERE tenant_id = 'default'
  AND (code IN ('OperationLog', 'LoginLog', 'JobLog')
       OR name IN ('操作日志', '登录日志', '任务日志', 'JobLog')
       OR component LIKE 'log/%'
       OR path IN ('/system/operation-log', '/system/login-log', '/system/job-log'))
  AND parent_id <> (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Log') t);