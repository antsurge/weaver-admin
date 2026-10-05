import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace PermissionRoleApi {
  // 角色
  export interface Role {
    id: string;
    name: string;
    code: string;
    weight: number;
    status: 'disabled' | 'enabled';
    remark: string;
    /** 绑定的数据权限规则ID列表 */
    dataPermissionIds?: string[];
    /** 是否系统内置 */
    isSystem?: boolean;
    /** 是否超级管理员角色 */
    isSuperAdmin?: boolean;
    /** 关联的菜单ID列表 */
    menuIds?: string[];
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface RoleListParams extends PaginationParams {
    name?: string;
    code?: string;
    status?: 'disabled' | 'enabled';
  }

  export interface isExists {
    exists: boolean;
  }
}

async function getRoleListApi(params?: PermissionRoleApi.RoleListParams) {
  return requestClient.get<PaginationResult<PermissionRoleApi.Role>>(
    '/admin/v1/role',
    {
      params,
    },
  );
}

async function getRoleApi(id: string) {
  return requestClient.get<PermissionRoleApi.Role>(`/admin/v1/role/${id}`);
}

async function createRoleApi(
  data: Omit<PermissionRoleApi.Role, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/role', data, {
    showSuccessMessage: true,
  });
}

async function updateRoleApi(
  id: string,
  data: Omit<PermissionRoleApi.Role, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/role/${id}`, data, {
    showSuccessMessage: true,
  });
}

async function updateRoleStatusApi(
  id: string,
  status: PermissionRoleApi.Role['status'],
) {
  return requestClient.put(`/admin/v1/role/${id}/status`, {
    status,
  });
}

async function deleteRoleApi(ids: string[]) {
  return requestClient.delete('/admin/v1/role', {
    params: {
      ids,
    },
    showSuccessMessage: true,
  });
}

/**
 * 为角色绑定菜单（全量替换）
 * @param roleId 角色ID
 * @param menuIds 菜单ID列表
 */
async function bindMenusForRoleApi(roleId: string, menuIds: string[]) {
  return requestClient.put(
    `/admin/v1/role/${roleId}/menus`,
    {
      menuIds,
    },
    {
      showSuccessMessage: false,
    },
  );
}

/**
 * 查询角色的菜单树
 * @param roleId 角色ID
 */
async function getMenusByRoleApi(roleId: string) {
  return requestClient.get(
    '/admin/v1/role/{roleId}/menus'.replace('{roleId}', roleId),
  );
}

/**
 * 为角色绑定数据权限规则（全量替换）
 * @param roleId 角色ID
 * @param dataPermissionIds 数据权限规则ID列表
 */
async function bindDataPermissionsForRoleApi(
  roleId: string,
  dataPermissionIds: string[],
) {
  return requestClient.put(
    `/admin/v1/role/${roleId}/data-permissions`,
    {
      dataPermissionIds,
    },
    {
      showSuccessMessage: false,
    },
  );
}

/**
 * 查询角色已绑定的数据权限规则列表
 * @param roleId 角色ID
 */
async function getDataPermissionsByRoleApi(roleId: string) {
  return requestClient.get(`/admin/v1/role/${roleId}/data-permissions`);
}

/**
 * 角色编码是否存在（用于表单失焦校验）
 * @param code 角色编码
 * @param id 角色ID（编辑时排除自身）
 */
async function isRoleCodeExistsApi(
  code: string,
  id?: PermissionRoleApi.Role['id'],
) {
  return requestClient.get<PermissionRoleApi.isExists>(
    '/admin/v1/role:code-exists',
    {
      params: { id, code },
    },
  );
}

export {
  bindDataPermissionsForRoleApi,
  bindMenusForRoleApi,
  createRoleApi,
  deleteRoleApi,
  getDataPermissionsByRoleApi,
  getMenusByRoleApi,
  getRoleApi,
  getRoleListApi,
  isRoleCodeExistsApi,
  updateRoleApi,
  updateRoleStatusApi,
};
