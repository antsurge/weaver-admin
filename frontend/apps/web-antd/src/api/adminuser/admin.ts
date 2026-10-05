import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

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
    /** 头像（对象键，由后端存储） */
    avatar?: string;
    /** 头像访问地址（由后端根据对象键生成，可直接用于 <img src>） */
    avatarUrl?: string;
    /** 状态 */
    status: 'disabled' | 'enabled';
    /**
     * 密码（创建时可选，留空使用后端默认密码；编辑时可选，留空不修改密码，非空则更新）
     */
    password?: string;
    /** 关联的角色ID列表 */
    roleIds?: string[];
    /** 关联的角色名称列表 */
    roleNames?: string[];
    /** 是否超级管理员（不可删除） */
    isSuperAdmin?: boolean;
    /** 归属部门ID */
    departmentId?: string;
    /** 归属部门名称 */
    departmentName?: string;
    /** 绑定的数据权限规则ID列表（空 = 跟随角色） */
    dataPermissionIds?: string[];
    /** 绑定的数据权限规则名称列表（用于列表展示；空 = 跟随角色） */
    dataPermissionNames?: string[];
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface AdminListParams extends PaginationParams {
    status?: 'disabled' | 'enabled';
    /** 用户名（模糊匹配） */
    username?: string;
    /** 真实姓名（模糊匹配） */
    realName?: string;
    /** 手机号（模糊匹配） */
    phone?: string;
    /** 邮箱（模糊匹配） */
    email?: string;
    /** 归属部门ID（精确匹配） */
    departmentId?: string;
  }

  export interface isExists {
    exists: boolean;
  }
}

async function getAdminListApi(params?: AdminuserAdminApi.AdminListParams) {
  return requestClient.get<PaginationResult<AdminuserAdminApi.Admin>>(
    '/admin/v1/admin',
    {
      params,
    },
  );
}

async function getAdminApi(id: string) {
  return requestClient.get<AdminuserAdminApi.Admin>(`/admin/v1/admin/${id}`);
}

async function createAdminApi(
  data: Omit<AdminuserAdminApi.Admin, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/admin', data, {
    showSuccessMessage: true,
  });
}

async function updateAdminApi(
  id: string,
  data: Omit<AdminuserAdminApi.Admin, 'createdAt' | 'id' | 'updatedAt'>,
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

/** 批量启用/禁用用户 */
async function batchUpdateAdminStatusApi(
  ids: string[],
  status: AdminuserAdminApi.Admin['status'],
) {
  return requestClient.put(
    '/admin/v1/admin/batch/status',
    { ids, status },
    { showSuccessMessage: true },
  );
}

/** 重置指定用户密码 */
async function resetPasswordApi(id: string, password: string) {
  return requestClient.put(
    `/admin/v1/admin/${id}/password`,
    { password },
    { showSuccessMessage: true },
  );
}

async function deleteAdminApi(ids: string[]) {
  return requestClient.delete('/admin/v1/admin', {
    params: {
      ids,
    },
    showSuccessMessage: true,
  });
}

/** 导出用户列表为 CSV（返回原始响应，便于从响应头解析文件名） */
async function exportAdminApi(params?: AdminuserAdminApi.AdminListParams) {
  return requestClient.get('/admin/v1/admin/export', {
    params,
    responseType: 'blob',
    responseReturn: 'raw',
    showFailMessage: false,
  });
}

/**
 * 用户名是否存在
 * @param username 用户名
 * @param id 用户ID（编辑时排除自身）
 */
async function isUsernameExistsApi(username: string, id?: string) {
  return requestClient.get<AdminuserAdminApi.isExists>(
    '/admin/v1/admin:username-exists',
    {
      params: { id, username },
    },
  );
}

export {
  batchUpdateAdminStatusApi,
  createAdminApi,
  deleteAdminApi,
  exportAdminApi,
  getAdminApi,
  getAdminListApi,
  isUsernameExistsApi,
  resetPasswordApi,
  updateAdminApi,
  updateAdminStatusApi,
};
