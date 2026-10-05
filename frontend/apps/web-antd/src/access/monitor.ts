/**
 * 系统监控模块权限标识
 * 与菜单管理中的 authCode 一一对应，供 v-access / hasAccessByCodes 使用
 */
export const MonitorAuthCode = {
  /** 系统监控（顶级目录） */
  ROOT: 'Monitor',
  Server: {
    /** 服务监控 */
    ROOT: 'Monitor:ServerMonitor',
  },
  Cache: {
    /** 缓存监控 */
    ROOT: 'Monitor:CacheMonitor',
  },
  Database: {
    /** 数据库监控 */
    ROOT: 'Monitor:DatabaseMonitor',
  },
  Api: {
    /** 接口监控 */
    ROOT: 'Monitor:ApiMonitor',
  },
  Job: {
    /** 定时任务监控 */
    ROOT: 'Monitor:JobMonitor',
  },
} as const;

export type MonitorAuthCodeValue =
  (typeof MonitorAuthCode)[keyof typeof MonitorAuthCode];
