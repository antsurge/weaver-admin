/**
 * 权限管理模块权限标识
 * 与菜单管理中的 authCode 一一对应，供 v-access / hasAccessByCodes 使用
 */
export const PermissionAuthCode = {
  /** 权限管理（顶级目录） */
  ROOT: 'Permission',
  DataPermission: {
    /** 数据权限 */
    ROOT: 'Permission:DataPermission',
    /** 按钮：查看详情 */
    Info: 'Permission:DataPermission:Info',
    /** 按钮：查看列表 */
    List: 'Permission:DataPermission:List',
    /** 按钮：新增 */
    Create: 'Permission:DataPermission:Create',
    /** 按钮：编辑 */
    Edit: 'Permission:DataPermission:Edit',
    /** 按钮：删除 */
    Delete: 'Permission:DataPermission:Delete',
    /** 按钮：批量删除 */
    BatchDelete: 'Permission:DataPermission:BatchDelete',
    /** 按钮：启用/禁用 */
    SwitchStatus: 'Permission:DataPermission:SwitchStatus',
  },
  Role: {
    /** 角色管理 */
    ROOT: 'Permission:Role',
    /** 按钮：查看详情 */
    Info: 'Permission:Role:Info',
    /** 按钮：查看列表 */
    List: 'Permission:Role:List',
    /** 按钮：新增 */
    Create: 'Permission:Role:Create',
    /** 按钮：编辑 */
    Edit: 'Permission:Role:Edit',
    /** 按钮：删除 */
    Delete: 'Permission:Role:Delete',
    /** 按钮：批量删除 */
    BatchDelete: 'Permission:Role:BatchDelete',
    /** 按钮：启用/禁用 */
    SwitchStatus: 'Permission:Role:SwitchStatus',
  },
  Menu: {
    /** 菜单管理 */
    ROOT: 'Permission:Menu',
    /** 按钮：查看详情 */
    Info: 'Permission:Menu:Info',
    /** 按钮：查看列表 */
    List: 'Permission:Menu:List',
    /** 按钮：启用/禁用 */
    SwitchStatus: 'Permission:Menu:SwitchStatus',
    /** 按钮：编辑 */
    Edit: 'Permission:Menu:Edit',
    /** 按钮：新增 */
    Create: 'Permission:Menu:Create',
    /** 按钮：新增下级 */
    AddSubordinate: 'Permission:Menu:AddSubordinate',
    /** 按钮：删除 */
    Delete: 'Permission:Menu:Delete',
    /** 按钮：批量删除 */
    BatchDelete: 'Permission:Menu:BatchDelete',
    /** 按钮：展开全部 */
    ExpandAll: 'Permission:Menu:ExpandAll',
    /** 按钮：折叠全部 */
    CollapseAll: 'Permission:Menu:CollapseAll',
  },
  Position: {
    /** 岗位管理 */
    ROOT: 'Organization:Position',
    /** 按钮：查看详情 */
    Info: 'Organization:Position:Info',
    /** 按钮：查看列表 */
    List: 'Organization:Position:List',
    /** 按钮：新增 */
    Create: 'Organization:Position:Create',
    /** 按钮：编辑 */
    Edit: 'Organization:Position:Edit',
    /** 按钮：删除 */
    Delete: 'Organization:Position:Delete',
    /** 按钮：批量删除 */
    BatchDelete: 'Organization:Position:BatchDelete',
    /** 按钮：导入 */
    Import: 'Organization:Position:Import',
    /** 按钮：导出 */
    Export: 'Organization:Position:Export',
    /** 按钮：下载导入模版 */
    DownloadTemplate: 'Organization:Position:DownloadTemplate',
  },
  Department: {
    /** 部门管理 */
    ROOT: 'Organization:Department',
    /** 按钮：查看详情 */
    Info: 'Organization:Department:Info',
    /** 按钮：查看列表 */
    List: 'Organization:Department:List',
    /** 按钮：新增 */
    Create: 'Organization:Department:Create',
    /** 按钮：编辑 */
    Edit: 'Organization:Department:Edit',
    /** 按钮：删除 */
    Delete: 'Organization:Department:Delete',
    /** 按钮：批量删除 */
    BatchDelete: 'Organization:Department:BatchDelete',
    /** 按钮：新增下级 */
    Append: 'Organization:Department:Append',
    /** 按钮：启用/禁用 */
    SwitchStatus: 'Organization:Department:SwitchStatus',
    /** 按钮：展开全部 */
    ExpandAll: 'Organization:Department:ExpandAll',
    /** 按钮：折叠全部 */
    CollapseAll: 'Organization:Department:CollapseAll',
  },
} as const;

export type PermissionAuthCodeValue =
  (typeof PermissionAuthCode)[keyof typeof PermissionAuthCode];
