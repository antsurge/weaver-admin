import { requestClient } from '#/api/request';
import type { PaginationParams,PaginationResult } from '#/types/pagination'

export namespace SystemPositionApi {
  /** 岗位 */
  export interface Position {
    /** 岗位ID */
    id: string;
    /** 岗位名称 */
    name: string;
    /** 岗位编码 */
    code: string;
    /** 权重 */
    weight: number;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'enabled' | 'disabled';
    /** 备注 */
    remark: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface PositionListParams extends PaginationParams{
    name?: string;
    code?: string;
    status?: 'enabled' | 'disabled';
  }
}

/**
 * 获取岗位列表
 */
async function getPositionListApi(params?:SystemPositionApi.PositionListParams) {
  return requestClient.get<PaginationResult<SystemPositionApi.Position>>(
    '/admin/v1/position',
    {
      params:params
    }
  );
}

/**
 * 获取岗位
 */
async function getPositionApi(id: string, params: object = {}) {
  return requestClient.get<SystemPositionApi.Position>(
    `/admin/v1/position/${id}`
  );
}

/**
 * 创建岗位
 * @param data 岗位数据
 */
async function createPositionApi(
  data: Omit<SystemPositionApi.Position, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/position', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新岗位
 * @param id 岗位 ID
 * @param data 岗位数据
 */
async function updatePositionApi(
  id: string,
  data: Omit<SystemPositionApi.Position, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/position/${id}`, data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新岗位状态
 * @param id 岗位 ID
 * @param status 状态
 */
async function updatePositionStatusApi(
  id: string,
  status: SystemPositionApi.Position['status'],
) {
  return requestClient.put(`/admin/v1/position/${id}/status`, {
    status,
  });
}

/**
 * 删除岗位
 * @param ids 岗位 ID 集合
 */
async function deletePositionApi(ids: string[]) {
  return requestClient.delete('/admin/v1/position', {
    params: {
      ids,
    },
    showSuccessMessage: true,
  });
}

export {
  getPositionListApi,
  getPositionApi,
  createPositionApi,
  updatePositionApi,
  updatePositionStatusApi,
  deletePositionApi,
};
