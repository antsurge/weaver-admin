-- =====================================================================
-- 低代码（Lowcode）菜单迁移脚本
-- 数据库：kratos_admin，表：menu
-- 目的：
--   1. 把原挂在"系统管理"下的"低代码"目录提升为顶级目录（parent_id=''，
--      path=/lowcode），成为独立菜单组。
--   2. 其子菜单（表单构建器/代码生成器）随之重挂到低代码目录下，
--      并修正 path/component/title 为规范值
--      （/system/lowcode/* -> /lowcode/*，system/* -> lowcode/*，
--       title 统一为 lowcode.* i18n key）。
-- 说明：
--   * 幂等：可重复执行；已迁移过的记录不会被重复挂载。
--   * 本脚本针对 default 租户（tenant_id='default'）生效；
--     若使用 multi 租户，请把 tenant_id 换成目标租户或去掉条件。
-- 执行方式：mysql -uroot kratos_admin < update-menu-lowcode.sql
-- =====================================================================

USE kratos_admin;

-- ---------------------------------------------------------------------
-- 1. 低代码目录提升为顶级目录（parent_id=''，path=/lowcode）
--    幂等：只有 parent_id 仍指向系统管理或 path 仍为旧值时更新
-- ---------------------------------------------------------------------
UPDATE menu
SET parent_id = '',
    name = '低代码',
    title = 'lowcode.title',
    path = '/lowcode',
    type = 'catalog',
    weight = 9
WHERE tenant_id = 'default'
  AND (code = 'Lowcode' OR name = '低代码' OR path = '/system/lowcode')
  AND (parent_id <> '' OR path <> '/lowcode');

-- ---------------------------------------------------------------------
-- 2. 子菜单重挂到低代码目录下，并修正 path/component/title
-- ---------------------------------------------------------------------

-- 2.1 表单构建器
UPDATE menu
SET parent_id = (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Lowcode') t),
    path = '/lowcode/form-builder',
    component = 'lowcode/form-builder/index',
    title = 'lowcode.formBuilder'
WHERE tenant_id = 'default'
  AND (code = 'FormBuilder' OR name = '表单构建器'
       OR path = '/system/lowcode/form-builder' OR component = 'system/form-builder/index')
  AND path <> '/lowcode/form-builder';

-- 2.2 代码生成器
UPDATE menu
SET parent_id = (SELECT id FROM (SELECT id FROM menu WHERE tenant_id = 'default' AND code = 'Lowcode') t),
    path = '/lowcode/codegen',
    component = 'lowcode/codegen/index',
    title = 'lowcode.codegen'
WHERE tenant_id = 'default'
  AND (code = 'Codegen' OR name = '代码生成器'
       OR path = '/system/lowcode/codegen' OR component = 'system/codegen/index')
  AND path <> '/lowcode/codegen';

-- ---------------------------------------------------------------------
-- 3. 验证：输出低代码目录及其子菜单
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
WHERE parent.code = 'Lowcode' AND child.tenant_id = 'default'
ORDER BY child.weight;

-- 仍使用旧路径/旧组件的低代码菜单（期望为空，若有说明 code/name 不匹配，可手动核对）
SELECT id, name, code, parent_id, path, component
FROM menu
WHERE tenant_id = 'default'
  AND (code = 'Lowcode' OR code = 'FormBuilder' OR code = 'Codegen'
       OR name IN ('低代码', '表单构建器', '代码生成器'))
  AND (path LIKE '/system/lowcode%'
       OR component IN ('system/form-builder/index', 'system/codegen/index')
       OR (code = 'Lowcode' AND parent_id <> ''));

-- ---------------------------------------------------------------------
-- 4. 修正 title：i18n key 统一为 lowcode.*
--    历史版本把 title 写成了 page.lowcode.*（词条不存在，菜单/页面国际化丢失），
--    现统一为 lowcode.title / lowcode.formBuilder / lowcode.codegen。
--    幂等：仅更新仍为旧 key 的记录。
-- ---------------------------------------------------------------------
UPDATE menu SET title = 'lowcode.title'       WHERE tenant_id = 'default' AND code = 'Lowcode'      AND title = 'page.lowcode.title';
UPDATE menu SET title = 'lowcode.formBuilder' WHERE tenant_id = 'default' AND code = 'FormBuilder' AND title = 'page.lowcode.formBuilder';
UPDATE menu SET title = 'lowcode.codegen'     WHERE tenant_id = 'default' AND code = 'Codegen'     AND title = 'page.lowcode.codegen';
