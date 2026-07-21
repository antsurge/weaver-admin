import { requestClient } from '#/api/request';
import type { PaginationParams, PaginationResult } from '#/types/pagination'

export namespace OrganizationPositionApi {
  /** 职务 */
  export interface Position {
    /** 职务ID */
    id: string;
    /** 职务名称 */
    name: string;
    /** 职务编码 */
    code: string;
    /** 权重/职务级别（越小职务越高） */
    weight: number;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'enabled' | 'disabled';
    /** 备注 */
    remark: string;
    /** 创建时间 */
    createdAt?: string;
    /** 创建人 */
    createdBy?: string;
    /** 创建人姓名 */
    createdByName?: string;
    /** 更新时间 */
    updatedAt?: string;
    /** 更新人 */
    updatedBy?: string;
    /** 更新人姓名 */
    updatedByName?: string;
  }

  /** 列表筛选参数 */
  export interface PositionListParams extends PaginationParams {
    /** 职务名称 */
    name?: string;
    /** 职务编码 */
    code?: string;
    /** 状态 */
    status?: 'enabled' | 'disabled';
  }

  export interface isExists {
    exists: boolean
  }
}

/**
 * 获取职务列表
 */
async function getPositionListApi(params?: OrganizationPositionApi.PositionListParams) {
  return requestClient.get<PaginationResult<OrganizationPositionApi.Position>>(
    '/admin/v1/position',
    {
      params: params
    }
  );
}

/**
 * 获取职务
 */
async function getPositionApi(id: string, params: object = {}) {
  return requestClient.get<OrganizationPositionApi.Position>(
    `/admin/v1/position/${id}`
  );
}

/**
 * 创建职务
 * @param data 职务数据
 */
async function createPositionApi(
  data: Omit<OrganizationPositionApi.Position, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/position', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新职务
 * @param id 职务 ID
 * @param data 职务数据
 */
async function updatePositionApi(
  id: string,
  data: Omit<OrganizationPositionApi.Position, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/position/${id}`, data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新职务状态
 * @param id 职务 ID
 * @param status 状态
 */
async function updatePositionStatusApi(
  id: string,
  status: OrganizationPositionApi.Position['status'],
) {
  return requestClient.put(`/admin/v1/position/${id}/status`, {
    status,
  });
}

/**
 * 删除职务
 * @param ids 职务 ID 集合
 */
async function deletePositionApi(ids: string[]) {
  return requestClient.delete('/admin/v1/position', {
    params: {
      ids,
    },
    showSuccessMessage: true,
  });
}

/**
 * 职位名称是否存在
 */
async function isPositionNameExistsApi(
  name: string,
  id?: OrganizationPositionApi.Position['id'],
) {
  return requestClient.get<OrganizationPositionApi.isExists>('/admin/v1/position:name-exists', {
    params: { id, name },
  });
}

/**
 * 职位编码是否存在
 */
async function isPositionCodeExistsApi(
  code: string,
  id?: OrganizationPositionApi.Position['id'],
) {
  return requestClient.get<OrganizationPositionApi.isExists>('/admin/v1/position:code-exists', {
    params: { id, code },
  });
}

/** 导出 */
async function exportPositionApi(params?: OrganizationPositionApi.PositionListParams) {
  return requestClient.post(
    '/admin/v1/position:export',
    params,
    {
      responseType: 'blob',
      responseReturn:"raw",
      showFailMessage:false
    },
  );
}

/** 导入 */
async function importPositionApi(data: FormData) {
  return requestClient.post('/admin/v1/position:import',data, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}



export {
  getPositionListApi,
  getPositionApi,
  createPositionApi,
  updatePositionApi,
  updatePositionStatusApi,
  deletePositionApi,
  isPositionNameExistsApi,
  isPositionCodeExistsApi,
  exportPositionApi,
  importPositionApi,
};
