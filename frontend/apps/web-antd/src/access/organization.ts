/**
 * 组织架构模块权限标识
 * 与菜单管理中的 authCode 一一对应，供 v-access / hasAccessByCodes 使用
 */
export const OrganizationAuthCode = {
  /** 组织架构（顶级目录） */
  ROOT: 'Organization',
  Department: {
    /** 部门管理 */
    ROOT: 'Organization:Department',
  },
  Position: {
    /** 岗位管理 */
    ROOT: 'Organization:Position',
  },
} as const;

export type OrganizationAuthCodeValue =
  (typeof OrganizationAuthCode)[keyof typeof OrganizationAuthCode];
