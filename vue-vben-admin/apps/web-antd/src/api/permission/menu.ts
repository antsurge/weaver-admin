import type { Recordable } from '@vben/types';
import type { AllResult } from '#/types/pagination'


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
  export interface PermissionMenu {
    [key: string]: any;
    /** 后端权限标识 */
    authCode: string;
    /** 子级 */
    children?: PermissionMenu[];
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

  export interface MenuTreeParams {
    name?: string
    cod?: string
  }
}

/**
 * 获取权限数据列表(Tree)
 */
async function getMenuTreeApi(params: PermissionMenuApi.MenuTreeParams = {}) {
  return requestClient.get<AllResult<PermissionMenuApi.PermissionMenu>>(
    '/admin/v1/menu/tree',
    {
      params: params
    }
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
 * @param data 权限数据
 */
async function updateMenuStatusApi(
  id: string,
  status: string,
) {
  return requestClient.put(`/admin/v1/menu/${id}/status`, {
    status
  });
}

/**
 * 删除权限
 * @param id 权限 ID
 */
async function deleteMenuApi(ids: string[]) {
  return requestClient.delete(`/admin/v1/menu`, {
    params: {
      ids: ids
    }
  });
}


// 检测code是否存在
async function isMenuCodeExists(
  name: string,
  id?: PermissionMenuApi.PermissionMenu['id'],
) {
  return requestClient.get<boolean>('/system/menu/name-exists', {
    params: { id, name },
  });
}

// 检测path是否存在
async function isMenuPathExists(
  path: string,
  id?: PermissionMenuApi.PermissionMenu['id'],
) {
  return requestClient.get<boolean>('/system/menu/path-exists', {
    params: { id, path },
  });
}

export {
  createMenuApi,
  deleteMenuApi,
  getMenuTreeApi,
  isMenuCodeExists,
  isMenuPathExists,
  updateMenuApi,
  updateMenuStatusApi,
};
