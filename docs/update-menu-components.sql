-- =====================================================================
-- 菜单 component 路径更新脚本（前端 views 目录按业务域重划分后）
-- 数据库：kratos_admin，表：menu
-- 说明：仅更新 component 字段，path/name 等保持不变
-- 执行方式：mysql -uroot kratos_admin < update-menu-components.sql
-- =====================================================================

USE kratos_admin;

-- 安全域（security）
UPDATE menu SET component = 'security/online/index'        WHERE component = 'system/online/index';
UPDATE menu SET component = 'security/settings/index'      WHERE component = 'system/security/index';

-- 日志域（log）：原安全域下的日志视图迁移到 views/log/ 后，component 同步更新
UPDATE menu SET component = 'log/login-log/index'          WHERE component = 'security/log/login-log/index';
UPDATE menu SET component = 'log/operation-log/index'      WHERE component = 'security/log/operation-log/index';

-- 消息域（message）
UPDATE menu SET component = 'message/notification/index'   WHERE component = 'system/notification/index';
UPDATE menu SET component = 'message/message-center/index' WHERE component = 'system/message-center/index';

-- 运维域（ops）
UPDATE menu SET component = 'ops/monitor/server/index'     WHERE component = 'system/monitor/server/index';
UPDATE menu SET component = 'ops/monitor/cache/index'      WHERE component = 'system/monitor/cache/index';
UPDATE menu SET component = 'ops/monitor/database/index'   WHERE component = 'system/monitor/database/index';
UPDATE menu SET component = 'ops/monitor/api/index'        WHERE component = 'system/monitor/api/index';
UPDATE menu SET component = 'ops/monitor/job/index'        WHERE component = 'system/monitor/job/index';
UPDATE menu SET component = 'ops/job/index'                WHERE component = 'system/job/index';
UPDATE menu SET component = 'ops/job-log/index'            WHERE component = 'system/job-log/index';

-- 租户域（tenant）
UPDATE menu SET component = 'tenant/index'                 WHERE component = 'system/tenant/index';

-- 权限域（permission）
UPDATE menu SET component = 'permission/data-permission/index' WHERE component = 'system/data-permission/index';

-- 低代码域（lowcode）
UPDATE menu SET component = 'lowcode/form-builder/index' WHERE component = 'system/form-builder/index';
UPDATE menu SET component = 'lowcode/codegen/index'     WHERE component = 'system/codegen/index';

-- 验证（应返回 0 行）
SELECT id, name, component FROM menu WHERE component LIKE 'system/%';
