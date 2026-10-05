import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace SystemApiInterfaceApi {
  export interface ApiInterface {
    id: string;
    service: string;
    tag: string;
    method: string;
    path: string;
    summary: string;
    code: string;
    createdAt?: string;
    updatedAt?: string;
  }

  export interface ApiInterfaceListParams extends PaginationParams {
    service?: string;
    tag?: string;
    method?: string;
    path?: string;
    summary?: string;
  }

  /** 新增/编辑请求体（code 由后端根据 service|method|path 自动生成） */
  export interface ApiInterfaceFormData {
    service: string;
    tag?: string;
    method: string;
    path: string;
    summary?: string;
  }

  export interface ImportResult {
    total: number;
    imported: number;
    updated: number;
    skipped: number;
  }

  export interface ApiInterfaceOptions {
    services: string[];
    tags: string[];
  }
}

// 分页查询接口列表
async function getApiInterfaceListApi(
  params?: SystemApiInterfaceApi.ApiInterfaceListParams,
) {
  return requestClient.get<
    PaginationResult<SystemApiInterfaceApi.ApiInterface>
  >('/admin/v1/api-interface', { params });
}

// 查询接口筛选选项（服务名/标签，下拉数据源）
async function getApiInterfaceOptionsApi() {
  return requestClient.get<SystemApiInterfaceApi.ApiInterfaceOptions>(
    '/admin/v1/api-interface/options',
  );
}

// 手动新增接口
async function createApiInterfaceApi(
  data: SystemApiInterfaceApi.ApiInterfaceFormData,
) {
  return requestClient.post<SystemApiInterfaceApi.ApiInterface>(
    '/admin/v1/api-interface',
    data,
  );
}

// 手动编辑接口
async function updateApiInterfaceApi(
  id: string,
  data: SystemApiInterfaceApi.ApiInterfaceFormData,
) {
  return requestClient.put<SystemApiInterfaceApi.ApiInterface>(
    `/admin/v1/api-interface/${id}`,
    data,
  );
}

// 导入 openapi.yaml 文件（按 code 对比：已存在更新、不存在创建）
async function importApiInterfaceApi(data: FormData) {
  return requestClient.post<SystemApiInterfaceApi.ImportResult>(
    '/admin/v1/api-interface/import',
    data,
    {
      headers: { 'Content-Type': 'multipart/form-data' },
    },
  );
}

// 批量删除接口
async function deleteApiInterfaceApi(ids: string[]) {
  return requestClient.delete('/admin/v1/api-interface', {
    params: { ids },
  });
}

export {
  createApiInterfaceApi,
  deleteApiInterfaceApi,
  getApiInterfaceListApi,
  getApiInterfaceOptionsApi,
  importApiInterfaceApi,
  updateApiInterfaceApi,
};
