/**
 * 管理员用户模块权限标识
 * 与菜单管理中的 authCode 一一对应，供 v-access / hasAccessByCodes 使用
 */
export const AdminAuthCode = {
  /** 管理员用户（顶级菜单） */
  ROOT: 'Admin',
  Admin: {
    /** 管理员用户 */
    ROOT: 'Adminuser:Admin',
    /** 按钮：查看详情 */
    Info: 'Adminuser:Admin:Info',
    /** 按钮：查看列表 */
    List: 'Adminuser:Admin:List',
    /** 按钮：新增 */
    Create: 'Adminuser:Admin:Create',
    /** 按钮：编辑 */
    Edit: 'Adminuser:Admin:Edit',
    /** 按钮：删除 */
    Delete: 'Adminuser:Admin:Delete',
    /** 按钮：批量删除 */
    BatchDelete: 'Adminuser:Admin:BatchDelete',
    /** 按钮：导出 */
    Export: 'Adminuser:Admin:Export',
    /** 按钮：重置密码 */
    ResetPassword: 'Adminuser:Admin:ResetPassword',
    /** 按钮：启用/禁用 */
    SwitchStatus: 'Adminuser:Admin:SwitchStatus',
  },
} as const;

export type AdminAuthCodeValue =
  (typeof AdminAuthCode)[keyof typeof AdminAuthCode];
