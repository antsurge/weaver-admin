import { requestClient } from '#/api/request';
import type { PaginationParams, PaginationResult } from '#/types/pagination'

export namespace AdminuserAdminApi {
  // 用户
  export interface Admin {
    /** 用户ID */
    id: string;
    /** 真实姓名 */
    realName: string;
    /** 用户名 */
    username: string;
    /** 邮箱 */
    email?: string;
    /** 手机号 */
    phone?: string;
    /** 头像 */
    avatar?: string;
    /** 状态 */
    status: 'enabled' | 'disabled';
    /**
     * 密码（创建时可选，留空使用后端默认密码；编辑时可选，留空不修改密码，非空则更新）
     */
    password?: string;
    /** 关联的角色ID列表 */
    roleIds?: string[];
    /** 创建时间 */
    createTime?: string;
    /** 更新时间 */
    updateTime?: string;
  }

  export interface AdminListParams extends PaginationParams {
    status?: 'enabled' | 'disabled';
  }
}

async function getAdminListApi(params?:AdminuserAdminApi.AdminListParams) {
  return requestClient.get<PaginationResult<AdminuserAdminApi.Admin>>(
    '/admin/v1/admin',
    {
      params:params
    }
  );
}

async function getAdminApi(id: string, params: object = {}) {
  return requestClient.get<AdminuserAdminApi.Admin>(
    `/admin/v1/admin/${id}`
  );
}

async function createAdminApi(
  data: Omit<AdminuserAdminApi.Admin, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/admin', data, {
    showSuccessMessage: true,
  });
}

async function updateAdminApi(
  id: string,
  data: Omit<AdminuserAdminApi.Admin, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/admin/${id}`, data, {
    showSuccessMessage: true,
  });
}

async function updateAdminStatusApi(
  id: string,
  status: AdminuserAdminApi.Admin['status'],
) {
  return requestClient.put(`/admin/v1/admin/${id}/status`, {
    status,
  });
}

async function deleteAdminApi(ids: string[]) {
  return requestClient.delete('/admin/v1/admin', {
    params: {
      ids,
    },
    showSuccessMessage: true,
  });
}


export {
  getAdminListApi,
  getAdminApi,
  createAdminApi,
  updateAdminApi,
  updateAdminStatusApi,
  deleteAdminApi,
}
