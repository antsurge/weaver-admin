import type { RouteRecordStringComponent } from '@vben/types';

import { requestClient } from '#/api/request';

/**
 * 获取用户所有菜单（旧接口，保留兼容）
 */
export async function getAllMenusApi() {
  return requestClient.get<RouteRecordStringComponent[]>('/menu/all');
}

/**
 * 获取当前登录用户的菜单（根据用户角色返回绑定的菜单树）
 */
export async function getCurrentUserMenusApi() {
  return requestClient.get('/admin/v1/current-user/menus');
}
