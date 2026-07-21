export interface PermissionItem {
  id: string
  name: string
  path: string
  parentId?: string
  icon?: string
  sort?: number
  status?: 0 | 1   // 0=禁用, 1=启用
  children?: PermissionItem[]
}
