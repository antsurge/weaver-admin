import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace SystemPermissionApi {
  /** 徽标颜色集合 */
  export const BadgeVariants = [
    'default',
    'destructive',
    'primary',
    'success',
    'warning',
  ] as const;
  /** 徽标类型集合 */
  export const BadgeTypes = ['dot', 'normal'] as const;
  /** 权限类型集合 */
  export const PermissionTypes = [
    'catalog',
    'menu',
    'embedded',
    'link',
    'button',
  ] as const;
  /** 系统权限 */
  export interface SystemPermission {
    [key: string]: any;
    /** 后端权限标识 */
    authCode: string;
    /** 子级 */
    children?: SystemPermission[];
    /** 组件 */
    component?: string;
    /** 权限ID */
    id: string;
    /** 权限元数据 */
    meta?: {
      /** 激活时显示的图标 */
      activeIcon?: string;
      /** 作为路由时，需要激活的权限的Path */
      activePath?: string;
      /** 固定在标签栏 */
      affixTab?: boolean;
      /** 在标签栏固定的顺序 */
      affixTabOrder?: number;
      /** 徽标内容(当徽标类型为normal时有效) */
      badge?: string;
      /** 徽标类型 */
      badgeType?: (typeof BadgeTypes)[number];
      /** 徽标颜色 */
      badgeVariants?: (typeof BadgeVariants)[number];
      /** 在权限中隐藏下级 */
      hideChildrenInMenu?: boolean;
      /** 在面包屑中隐藏 */
      hideInBreadcrumb?: boolean;
      /** 在权限中隐藏 */
      hideInMenu?: boolean;
      /** 在标签栏中隐藏 */
      hideInTab?: boolean;
      /** 权限图标 */
      icon?: string;
      /** 内嵌Iframe的URL */
      iframeSrc?: string;
      /** 是否缓存页面 */
      keepAlive?: boolean;
      /** 外链页面的URL */
      link?: string;
      /** 同一个路由最大打开的标签数 */
      maxNumOfOpenTab?: number;
      /** 无需基础布局 */
      noBasicLayout?: boolean;
      /** 是否在新窗口打开 */
      openInNewWindow?: boolean;
      /** 权限排序 */
      order?: number;
      /** 额外的路由参数 */
      query?: Recordable<any>;
      /** 权限标题 */
      title?: string;
    };
    /** 权限名称 */
    name: string;
    /** 路由路径 */
    path: string;
    /** 父级ID */
    pid: string;
    /** 重定向 */
    redirect?: string;
    /** 权限类型 */
    type: (typeof MenuTypes)[number];
  }
}

/**
 * 获取权限数据列表(Tree)
 */
async function getPermissionTreeApi() {
  return requestClient.get<Array<SystemPermissionApi.SystemPermission>>(
    '/admin/v1/permission/tree',
  );
}

/**
 * 创建权限
 * @param data 权限数据
 */
async function createPermission(
  data: Omit<SystemPermissionApi.SystemPermission, 'children' | 'id'>,
) {
  return requestClient.post('/admin/v1/permission', data);
}

/**
 * 更新权限
 *
 * @param id 权限 ID
 * @param data 权限数据
 */
async function updatePermission(
  id: string,
  data: Omit<SystemPermissionApi.SystemPermission, 'children' | 'id'>,
) {
  return requestClient.put(`/admin/v1/permission/${id}`, data);
}

/**
 * 更新权限状态
 *
 * @param id 权限 ID
 * @param data 权限数据
 */
async function updatePermissionStatusApi(
  id: string,
  status: string,
) {
  return requestClient.put(`/admin/v1/permission/${id}/status`, {
    status
  });
}

/**
 * 删除权限
 * @param id 权限 ID
 */
async function deletePermissionApi(ids: string[]) {
  return requestClient.delete(`/admin/v1/permission`, {
    params: {
      ids: ids
    }
  });
}


// 检测code是否存在
async function isPermissionCodeExists(
  name: string,
  id?: SystemPermissionApi.SystemPermission['id'],
) {
  return requestClient.get<boolean>('/system/permission/name-exists', {
    params: { id, name },
  });
}

// 检测path是否存在
async function isPermissionPathExists(
  path: string,
  id?: SystemPermissionApi.SystemPermission['id'],
) {
  return requestClient.get<boolean>('/system/permission/path-exists', {
    params: { id, path },
  });
}

export {
  createPermission,
  deletePermissionApi,
  getPermissionTreeApi,
  isPermissionCodeExists,
  isPermissionPathExists,
  updatePermission,
  updatePermissionStatusApi,
};
