import { requestClient } from '#/api/request';
import type { PaginationParams, PaginationResult } from '#/types/pagination';

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

  export interface ImportResult {
    total: number;
    imported: number;
    skipped: number;
  }
}

// 分页查询接口列表
async function getApiInterfaceListApi(
  params?: SystemApiInterfaceApi.ApiInterfaceListParams,
) {
  return requestClient.get<PaginationResult<SystemApiInterfaceApi.ApiInterface>>(
    '/admin/v1/api-interface',
    { params },
  );
}

// 导入 openapi.yaml 文件
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

export { getApiInterfaceListApi, importApiInterfaceApi, deleteApiInterfaceApi };