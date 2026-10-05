/**
 * 权限标识统一出口
 * 按模块组织（system/lowcode/monitor/security/logs/admin/organization/permission），
 * 每个模块导出嵌套对象，形如：{ Menu: { List: 'Permission:Menu:List' } }，
 * 与菜单管理中的 authCode 一一对应，变量全局唯一，避免硬编码字符串散落各处。
 *
 * 用法示例：
 * - 模板：<Button v-access:code="[AdminAuthCode.Admin.Create]">新增</Button>
 * - 逻辑：hasAccessByCodes([PermissionAuthCode.Menu.Create])
 */
import { AdminAuthCode } from './admin';
import { LogsAuthCode } from './logs';
import { LowcodeAuthCode } from './lowcode';
import { MonitorAuthCode } from './monitor';
import { OrganizationAuthCode } from './organization';
import { PermissionAuthCode } from './permission';
import { SecurityAuthCode } from './security';
import { SystemAuthCode } from './system';

export * from './admin';
export * from './logs';
export * from './lowcode';
export * from './monitor';
export * from './organization';
export * from './permission';
export * from './security';
export * from './system';

/** 递归收集嵌套对象中的所有叶子字符串值（即全部 authCode） */
function collectLeafValues(
  obj: Record<string, unknown>,
  acc: string[] = [],
): string[] {
  Object.values(obj).forEach((value) => {
    if (typeof value === 'string') {
      acc.push(value);
    } else if (value && typeof value === 'object') {
      collectLeafValues(value as Record<string, unknown>, acc);
    }
  });
  return acc;
}

/** 全量权限标识（含菜单 + 按钮），可用于全量比对 / 调试 */
export const AllAuthCodes: Readonly<string[]> = Object.freeze(
  collectLeafValues({
    ...SystemAuthCode,
    ...LowcodeAuthCode,
    ...MonitorAuthCode,
    ...SecurityAuthCode,
    ...LogsAuthCode,
    ...AdminAuthCode,
    ...OrganizationAuthCode,
    ...PermissionAuthCode,
  }),
);
