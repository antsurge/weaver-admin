import { requestClient } from '#/api/request';
import type { PaginationParams, PaginationResult } from '#/types/pagination'

export namespace PermissionRoleApi {
  // 角色
  export interface Role {
    id: string;
    name: string,
    code: string;
    weight: number;
    status: 'enabled' | 'disabled';
    remark: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface RoleListParams extends PaginationParams {
    name?: string;
    code?: string;
    status?: 'enabled' | 'disabled';
  }
}

async function getRoleListApi(params?:PermissionRoleApi.RoleListParams) {
  return requestClient.get<PaginationResult<PermissionRoleApi.Role>>(
    '/admin/v1/role',
    {
      params:params
    }
  );
}

async function getRoleApi(id: string, params: object = {}) {
  return requestClient.get<PermissionRoleApi.Role>(
    `/admin/v1/role/${id}`
  );
}

async function createRoleApi(
  data: Omit<PermissionRoleApi.Role, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/role', data, {
    showSuccessMessage: true,
  });
}

async function updateRoleApi(
  id: string,
  data: Omit<PermissionRoleApi.Role, 'id' | 'createdAt' | 'updatedAt'>,
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


export {
  getRoleListApi,
  getRoleApi,
  createRoleApi,
  updateRoleApi,
  updateRoleStatusApi,
  deleteRoleApi,
}
