import type { AllResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace PermissionMenuApi {
  /** 徽标颜色集合 */
  export const BadgeVariants = [
    'default',
    'destructive',
    'primary',
    'success',
    'warning',
  ] as const;
  /** 徽标类型集合 */
  export const BadgeTypes = ['dot', 'text'] as const;
  /** 权限类型集合 */
  export const PermissionTypes = [
    'catalog',
    'menu',
    'action',
    'iframe',
    'link',
  ] as const;
  /** 单条接口权限（按钮绑定）
   * 绑定基于 api_interface.code（service|METHOD|path）关联，接口信息实时反查自 api_interface
   */
  export interface ApiPermission {
    /** 服务全限定名 */
    service: string;
    /** OpenAPI 标签（前端分组回显用） */
    tag?: string;
    /** HTTP method，如 GET/POST */
    method: string;
    /** 接口路径 */
    path: string;
    /** 接口描述 */
    summary?: string;
    /** 业务唯一键 service|METHOD|path（对应 api_interface.code） */
    code?: string;
  }
  /** 单条接口端点 */
  export interface ApiEndpoint {
    /** HTTP method，如 GET/POST */
    method: string;
    /** 接口路径，如 /admin/v1/menu */
    path: string;
    /** 接口描述 */
    summary: string;
  }
  /** 按 service 分组的接口元数据 */
  export interface ApiMetadata {
    /** 服务名，如 PermissionService */
    service: string;
    /** OpenAPI tag */
    tag: string;
    endpoints: ApiEndpoint[];
  }
  /** 系统权限（字段与后端 Menu message 一致，均为扁平结构） */
  export interface PermissionMenu {
    [key: string]: any;
    /** 按钮绑定的接口权限列表（仅 action 类型） */
    apiPermissions?: ApiPermission[];
    /** 权限标识 */
    authCode: string;
    /** 徽标内容 */
    badge?: string;
    /** 徽标类型 */
    badgeType?: (typeof BadgeTypes)[number];
    /** 徽标样式 */
    badgeVariants?: (typeof BadgeVariants)[number];
    /** 子级 */
    children?: PermissionMenu[];
    /** 页面组件 */
    component?: string;
    /** 描述 */
    description?: string;
    /** 权限ID */
    id: string;
    /** 图标 */
    icon?: string;
    /** 链接地址（iframe/外链） */
    linkUrl?: string;
    /** 权限名称 */
    name: string;
    /** 路由路径 */
    path: string;
    /** 父级ID */
    parentID: string;
    /** 备注 */
    remark?: string;
    /** 状态 */
    status?: string;
    /** 标题（国际化 key） */
    title?: string;
    /** 权限类型 */
    type: (typeof PermissionTypes)[number];
    /** 权重 */
    weight?: number;
  }

  export interface MenuTreeParams {
    name?: string;
    status?: string;
    type?: string;
  }
}

/**
 * 获取权限数据列表(Tree)
 */
async function getMenuTreeApi(params: PermissionMenuApi.MenuTreeParams = {}) {
  return requestClient.get<AllResult<PermissionMenuApi.PermissionMenu>>(
    '/admin/v1/menu/tree',
    {
      params,
    },
  );
}

/**
 * 创建权限
 * @param data 权限数据
 */
async function createMenuApi(
  data: Omit<PermissionMenuApi.PermissionMenu, 'children' | 'id'>,
) {
  return requestClient.post('/admin/v1/menu', data);
}

/**
 * 获取权限详情（含接口权限）
 *
 * @param id 权限 ID
 */
async function getMenuDetailApi(id: string) {
  return requestClient.get<PermissionMenuApi.PermissionMenu>(
    `/admin/v1/menu/${id}`,
  );
}

/**
 * 更新权限
 *
 * @param id 权限 ID
 * @param data 权限数据
 */
async function updateMenuApi(
  id: string,
  data: Omit<PermissionMenuApi.PermissionMenu, 'children' | 'id'>,
) {
  return requestClient.put(`/admin/v1/menu/${id}`, data);
}

/**
 * 更新权限状态
 *
 * @param id 权限 ID
 * @param status 权限状态
 */
async function updateMenuStatusApi(id: string, status: string) {
  return requestClient.put(`/admin/v1/menu/${id}/status`, {
    status,
  });
}

/**
 * 删除权限
 * @param ids 权限 ID 列表
 */
async function deleteMenuApi(ids: string[]) {
  return requestClient.delete(`/admin/v1/menu`, {
    params: {
      ids,
    },
  });
}

// ========== 接口元数据（按钮绑定 API 用） ==========

/**
 * 列出所有接口元数据（按 service 分组）
 */
async function listApiMetadataApi(): Promise<PermissionMenuApi.ApiMetadata[]> {
  const res = await requestClient.get<{
    items: PermissionMenuApi.ApiMetadata[];
  }>('/admin/v1/api-metadata');
  return res?.items ?? [];
}

export {
  createMenuApi,
  deleteMenuApi,
  getMenuDetailApi,
  getMenuTreeApi,
  listApiMetadataApi,
  updateMenuApi,
  updateMenuStatusApi,
};
