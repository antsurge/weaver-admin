import { requestClient } from '#/api/request';
import type { PaginationParams,PaginationResult } from '#/types/pagination'

export namespace DictionaryDictTypeApi {
  /** 字典类型 */
  export interface DictType {
    /** 类型ID */
    id: string;
    /** 类型名称 */
    name: string;
    /** 类型编码 */
    code: string;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'enabled' | 'disabled';
    /** 备注 */
    remark?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface DictTypeListParams extends PaginationParams{

  }
}

/**
 * 获取字典类型列表
 */
async function getDictTypeListApi(params?:DictionaryDictTypeApi.DictTypeListParams) {
  return requestClient.get<PaginationResult<DictionaryDictTypeApi.DictType>>('/admin/v1/dict-type',{
    params:params
  });
}

/**
 * 获取字典类型
 */
async function getDictTypeApi(id:string) {
  return requestClient.get<DictionaryDictTypeApi.DictType>(
    `/admin/v1/dict-type/${id}`,
  );
}

/**
 * 创建字典类型
 */
async function createDictTypeApi(
  data: Omit<DictionaryDictTypeApi.DictType, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/dict-type', data,{
    showSuccessMessage: true,
  });
}

/**
 * 更新字典类型
 */
async function updateDictTypeApi(
  id: string,
  data: Omit<DictionaryDictTypeApi.DictType, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/dict-type/${id}`, data,{
    showSuccessMessage: true,
  });
}

/**
 * 更新字典类型状态
 */
async function updateDictTypeStatusApi(
  id: string,
  status: DictionaryDictTypeApi.DictType['status'],
) {
  return requestClient.put(`/admin/v1/dict-type/${id}/status`, {
    status,
  });
}

/**
 * 删除字典类型
 */
async function deleteDictTypeApi(ids: string[]) {
  return requestClient.delete('/admin/v1/dict-type', {
    params: {
      ids,
    },
  });
}

export {
  getDictTypeListApi,
  getDictTypeApi,
  createDictTypeApi,
  updateDictTypeApi,
  updateDictTypeStatusApi,
  deleteDictTypeApi,
};

