/**
 * 系统管理模块权限标识
 * 与菜单管理中的 authCode 一一对应，供 v-access / hasAccessByCodes 使用
 */
export const SystemAuthCode = {
  /** 系统管理（顶级目录） */
  ROOT: 'System',
  Tenant: {
    /** 租户管理 */
    ROOT: 'System:Tenant',
    /** 按钮：查看列表 */
    List: 'System:Tenant:List',
    /** 按钮：查看详情 */
    Info: 'System:Tenant:Info',
    /** 按钮：新增 */
    Create: 'System:Tenant:Create',
    /** 按钮：编辑 */
    Edit: 'System:Tenant:Edit',
    /** 按钮：删除 */
    Delete: 'System:Tenant:Delete',
    /** 按钮：启用/禁用 */
    SwitchStatus: 'System:Tenant:SwitchStatus',
  },
  ApiInterface: {
    /** 接口管理 */
    ROOT: 'System:ApiInterface',
    /** 按钮：查看列表 */
    List: 'System:ApiInterface:List',
    /** 按钮：新增 */
    Create: 'System:ApiInterface:Create',
    /** 按钮：编辑 */
    Edit: 'System:ApiInterface:Edit',
    /** 按钮：删除 */
    Delete: 'System:ApiInterface:Delete',
    /** 按钮：导入 */
    Import: 'System:ApiInterface:Import',
  },
  Notification: {
    /** 消息通知 */
    ROOT: 'System:Notification',
    /** 按钮：查看列表 */
    List: 'System:Notification:List',
    /** 按钮：发布 */
    Create: 'System:Notification:Create',
    /** 按钮：撤回 */
    Revoke: 'System:Notification:Revoke',
  },
  MessageCenter: {
    /** 消息中心 */
    ROOT: 'System:MessageCenter',
  },
  DictType: {
    /** 字典管理 */
    ROOT: 'System:DictType',
    /** 按钮：查看列表 */
    List: 'System:DictType:List',
    /** 按钮：查看详情 */
    Info: 'System:DictType:Info',
    /** 按钮：新增 */
    Create: 'System:DictType:Create',
    /** 按钮：编辑 */
    Edit: 'System:DictType:Edit',
    /** 按钮：删除 */
    Delete: 'System:DictType:Delete',
    /** 按钮：新增字典数据（新增下级） */
    AddSubordinate: 'System:DictType:AddSubordinate',
  },
  Config: {
    /** 参数设置 */
    ROOT: 'System:Config',
    /** 按钮：查看列表 */
    List: 'System:Config:List',
    /** 按钮：查看详情 */
    Info: 'System:Config:Info',
    /** 按钮：新增 */
    Create: 'System:Config:Create',
    /** 按钮：编辑 */
    Edit: 'System:Config:Edit',
    /** 按钮：删除 */
    Delete: 'System:Config:Delete',
    /** 按钮：批量删除 */
    BatchDelete: 'System:Config:BatchDelete',
    /** 按钮：启用/禁用 */
    SwitchStatus: 'System:Config:SwitchStatus',
  },
  Job: {
    /** 定时任务 */
    ROOT: 'System:Job',
    /** 按钮：查看列表 */
    List: 'System:Job:List',
    /** 按钮：查看详情 */
    Info: 'System:Job:Info',
    /** 按钮：新增 */
    Create: 'System:Job:Create',
    /** 按钮：编辑 */
    Edit: 'System:Job:Edit',
    /** 按钮：删除 */
    Delete: 'System:Job:Delete',
    /** 按钮：批量删除 */
    BatchDelete: 'System:Job:BatchDelete',
    /** 按钮：启用/禁用 */
    SwitchStatus: 'System:Job:SwitchStatus',
  },
} as const;

export type SystemAuthCodeValue =
  (typeof SystemAuthCode)[keyof typeof SystemAuthCode];
