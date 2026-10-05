import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace SystemConfigApi {
  /** 系统参数 */
  export interface Config {
    /** 参数ID */
    id: string;
    /** 参数名称 */
    name: string;
    /** 参数键名 */
    key: string;
    /** 参数键值 */
    value: string;
    /** 系统内置：Y=内置，N=外置 */
    configType: 'N' | 'Y';
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'disabled' | 'enabled';
    /** 备注 */
    remark?: string;
    /** 分组名称（可为空，即单一参数） */
    group?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface ConfigListParams extends PaginationParams {
    /** 参数名称（模糊） */
    name?: string;
    /** 参数键名（模糊） */
    key?: string;
    /** 状态 */
    status?: string;
    /** 分组名称 */
    group?: string;
  }
}

/**
 * 获取参数列表
 */
async function getConfigListApi(params?: SystemConfigApi.ConfigListParams) {
  return requestClient.get<PaginationResult<SystemConfigApi.Config>>(
    '/admin/v1/config',
    { params },
  );
}

/**
 * 根据参数键名获取参数值（供业务页面消费）
 */
async function getConfigByKeyApi(key: string) {
  return requestClient.get<SystemConfigApi.Config>(
    `/admin/v1/config/${key}/value`,
  );
}

/**
 * 创建参数
 */
async function createConfigApi(
  data: Omit<SystemConfigApi.Config, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/config', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新参数
 */
async function updateConfigApi(
  id: string,
  data: Omit<SystemConfigApi.Config, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/config/${id}`, data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新参数状态
 */
async function updateConfigStatusApi(
  id: string,
  status: SystemConfigApi.Config['status'],
) {
  return requestClient.put(`/admin/v1/config/${id}/status`, { status });
}

/**
 * 删除参数（批量）
 */
async function deleteConfigApi(ids: string[]) {
  return requestClient.delete('/admin/v1/config', {
    params: { ids },
  });
}

export {
  createConfigApi,
  deleteConfigApi,
  getConfigByKeyApi,
  getConfigListApi,
  updateConfigApi,
  updateConfigStatusApi,
};
