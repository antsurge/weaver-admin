import { requestClient } from '#/api/request';

/**
 * 获取当前登录用户的菜单（根据用户角色返回绑定的菜单树）
 */
export async function getCurrentUserMenusApi() {
  return requestClient.get('/admin/v1/current-user/menus');
}
