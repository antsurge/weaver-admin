/**
 * 安全中心模块权限标识
 * 与菜单管理中的 authCode 一一对应，供 v-access / hasAccessByCodes 使用
 */
export const SecurityAuthCode = {
  /** 安全中心（顶级目录） */
  ROOT: 'Security',
  Online: {
    /** 在线用户 */
    ROOT: 'Security:Online',
    /** 强制下线 */
    ForceLogout: 'Security:Online:ForceLogout',
  },
  Policy: {
    /** 安全策略 */
    ROOT: 'Security:SecurityPolicy',
    /** 编辑安全策略 */
    Edit: 'Security:SecurityPolicy:Edit',
  },
} as const;

export type SecurityAuthCodeValue =
  (typeof SecurityAuthCode)[keyof typeof SecurityAuthCode];
