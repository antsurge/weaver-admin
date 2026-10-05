import type { PaginationParams } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace DictionaryDictDataApi {
  /** 字典数据 */
  export interface DictData {
    /** 数据ID */
    id: string;
    /** 所属字典类型ID */
    dictTypeID: string;
    /** 显示标签 */
    label: string;
    /** 实际值 */
    value: string;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'disabled' | 'enabled';
    /** 备注 */
    remark?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface DictDataListParams extends PaginationParams {
    dictTypeID?: string;
    /** 字典类型ID列表（多选过滤） */
    dictTypeIds?: string[];
  }
}

/**
 * 根据字典类型获取字典数据列表
 */
async function getDictDataListApi(
  params?: DictionaryDictDataApi.DictDataListParams,
) {
  return requestClient.get<Array<DictionaryDictDataApi.DictData>>(
    '/admin/v1/dict-data',
    {
      params,
    },
  );
}

/**
 * 创建字典数据
 */
async function createDictDataApi(
  data: Omit<DictionaryDictDataApi.DictData, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/dict-data', data);
}

/**
 * 更新字典数据
 */
async function updateDictDataApi(
  id: string,
  data: Omit<DictionaryDictDataApi.DictData, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/dict-data/${id}`, data);
}

/**
 * 更新字典数据状态
 */
async function updateDictDataStatusApi(
  id: string,
  status: DictionaryDictDataApi.DictData['status'],
) {
  return requestClient.put(`/admin/v1/dict-data/${id}/status`, {
    status,
  });
}

/**
 * 删除字典数据
 */
async function deleteDictDataApi(ids: string[]) {
  return requestClient.delete('/admin/v1/dict-data', {
    params: {
      ids,
    },
  });
}

export {
  createDictDataApi,
  deleteDictDataApi,
  getDictDataListApi,
  updateDictDataApi,
  updateDictDataStatusApi,
};
