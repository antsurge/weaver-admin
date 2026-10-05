import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace SystemTenantApi {
  /** 租户 */
  export interface Tenant {
    /** ID */
    id: string;
    /** 租户编码（登录用，创建后不可修改） */
    code: string;
    /** 租户名称 */
    name: string;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'disabled' | 'enabled';
    /** 到期时间 */
    expireAt?: string;
    /** 最大用户数（0=不限） */
    maxUsers?: number;
    /** 最大角色数（0=不限） */
    maxRoles?: number;
    /** 联系人 */
    contactName?: string;
    /** 联系电话 */
    contactPhone?: string;
    /** 备注 */
    remark?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface TenantListParams extends PaginationParams {
    name?: string;
    code?: string;
    status?: string;
  }
}

/**
 * 获取租户列表
 */
async function getTenantListApi(params?: SystemTenantApi.TenantListParams) {
  return requestClient.get<PaginationResult<SystemTenantApi.Tenant>>(
    '/admin/v1/tenant',
    { params },
  );
}

/**
 * 获取租户详情
 */
async function getTenantApi(id: string) {
  return requestClient.get<SystemTenantApi.Tenant>(`/admin/v1/tenant/${id}`);
}

/**
 * 创建租户
 */
async function createTenantApi(
  data: Omit<SystemTenantApi.Tenant, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/tenant', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新租户
 */
async function updateTenantApi(
  id: string,
  data: Omit<SystemTenantApi.Tenant, 'code' | 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/tenant/${id}`, data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新租户状态
 */
async function updateTenantStatusApi(
  id: string,
  status: SystemTenantApi.Tenant['status'],
) {
  return requestClient.put(`/admin/v1/tenant/${id}/status`, { status });
}

/**
 * 删除租户（批量）
 */
async function deleteTenantApi(ids: string[]) {
  return requestClient.delete('/admin/v1/tenant', {
    params: {
      ids,
    },
  });
}

export {
  createTenantApi,
  deleteTenantApi,
  getTenantApi,
  getTenantListApi,
  updateTenantApi,
  updateTenantStatusApi,
};
