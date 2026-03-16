import { requestClient } from '#/api/request';

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
    /** 描述 */
    description?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }
}

/**
 * 获取岗位列表
 */
async function getPositionListApi() {
  return requestClient.get<Array<SystemPositionApi.Position>>(
    '/admin/v1/position',
  );
}

/**
 * 创建岗位
 * @param data 岗位数据
 */
async function createPositionApi(
  data: Omit<SystemPositionApi.Position, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/position', data);
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
  return requestClient.put(`/admin/v1/position/${id}`, data);
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
  });
}

export {
  createPositionApi,
  deletePositionApi,
  getPositionListApi,
  updatePositionApi,
  updatePositionStatusApi,
};
