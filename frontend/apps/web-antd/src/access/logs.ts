/**
 * 日志中心模块权限标识
 * 与菜单管理中的 authCode 一一对应，供 v-access / hasAccessByCodes 使用
 */
export const LogsAuthCode = {
  /** 日志中心（顶级目录） */
  ROOT: 'Logs',
  Operation: {
    /** 操作日志 */
    ROOT: 'Logs:OperationLog',
    /** 按钮：查看列表 */
    List: 'Logs:OperationLog:List',
    /** 按钮：查看详情 */
    Info: 'Logs:OperationLog:Info',
    /** 按钮：删除 */
    Delete: 'Logs:OperationLog:Delete',
    /** 按钮：清空 */
    Clear: 'Logs:OperationLog:Clear',
  },
  Login: {
    /** 登录日志 */
    ROOT: 'Logs:LoginLog',
    /** 按钮：查看列表 */
    List: 'Logs:LoginLog:List',
    /** 按钮：查看详情 */
    Info: 'Logs:LoginLog:Info',
    /** 按钮：删除 */
    Delete: 'Logs:LoginLog:Delete',
    /** 按钮：清空 */
    Clear: 'Logs:LoginLog:Clear',
  },
  Job: {
    /** 任务日志 */
    ROOT: 'Logs:JobLog',
    /** 按钮：查看列表 */
    List: 'Logs:JobLog:List',
    /** 按钮：查看详情 */
    Info: 'Logs:JobLog:Info',
    /** 按钮：删除 */
    Delete: 'Logs:JobLog:Delete',
    /** 按钮：清空 */
    Clear: 'Logs:JobLog:Clear',
  },
} as const;

export type LogsAuthCodeValue =
  (typeof LogsAuthCode)[keyof typeof LogsAuthCode];
