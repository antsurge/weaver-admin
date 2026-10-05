import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace SystemDataPermissionApi {
  /** 数据权限规则 */
  export interface DataPermission {
    /** 规则ID */
    id: string;
    /** 规则名称 */
    name: string;
    /** 规则编码 */
    code: string;
    /** 数据范围：1=全部数据 2=自定义 3=本部门 4=本部门及以下 5=仅本人 */
    scopeType: string;
    /** 自定义数据范围的部门ID集合（JSON数组） */
    deptIds?: string;
    /** 自定义数据范围的角色ID集合（JSON数组） */
    roleIds?: string;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'disabled' | 'enabled';
    /** 备注 */
    remark?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface DataPermissionListParams extends PaginationParams {
    /** 规则名称（模糊） */
    name?: string;
    /** 规则编码（模糊） */
    code?: string;
    /** 数据范围 */
    scopeType?: string;
    /** 状态 */
    status?: string;
  }

  export interface DataPermissionInput {
    name: string;
    code: string;
    scopeType: string;
    deptIds?: string;
    roleIds?: string;
    status: 'disabled' | 'enabled';
    remark?: string;
  }

  /** 编码是否存在响应 */
  export interface isExists {
    /** 是否存在 */
    exists: boolean;
  }
}

/**
 * 获取数据权限规则列表
 */
async function getDataPermissionListApi(
  params?: SystemDataPermissionApi.DataPermissionListParams,
) {
  return requestClient.get<
    PaginationResult<SystemDataPermissionApi.DataPermission>
  >('/admin/v1/data-permission', { params });
}

/**
 * 创建数据权限规则
 */
async function createDataPermissionApi(
  data: SystemDataPermissionApi.DataPermissionInput,
) {
  return requestClient.post('/admin/v1/data-permission', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新数据权限规则
 */
async function updateDataPermissionApi(
  id: string,
  data: SystemDataPermissionApi.DataPermissionInput,
) {
  return requestClient.put(`/admin/v1/data-permission/${id}`, data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新数据权限规则状态
 */
async function updateDataPermissionStatusApi(
  id: string,
  status: SystemDataPermissionApi.DataPermission['status'],
) {
  return requestClient.put(`/admin/v1/data-permission/${id}/status`, {
    status,
  });
}

/**
 * 删除数据权限规则（批量）
 */
async function deleteDataPermissionApi(ids: string[]) {
  return requestClient.delete('/admin/v1/data-permission', {
    params: { ids },
  });
}

/**
 * 规则编码是否存在（用于表单失焦校验）
 * @param code 规则编码
 * @param id 规则ID（编辑时排除自身）
 */
async function isDataPermissionCodeExistsApi(
  code: string,
  id?: SystemDataPermissionApi.DataPermission['id'],
) {
  return requestClient.get<SystemDataPermissionApi.isExists>(
    '/admin/v1/data-permission:code-exists',
    {
      params: { code, id },
    },
  );
}

export {
  createDataPermissionApi,
  deleteDataPermissionApi,
  getDataPermissionListApi,
  isDataPermissionCodeExistsApi,
  updateDataPermissionApi,
  updateDataPermissionStatusApi,
};
