import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

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
    status: 'disabled' | 'enabled';
    /** 备注 */
    remark?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
    /** 字典数据列表 */
    dictData?: DictDataItem[];
  }

  /** 字典数据（按编码查询时随类型一起返回） */
  export interface DictDataItem {
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

  export type DictTypeListParams = PaginationParams;
}

/**
 * 获取字典类型列表
 */
async function getDictTypeListApi(
  params?: DictionaryDictTypeApi.DictTypeListParams,
) {
  return requestClient.get<PaginationResult<DictionaryDictTypeApi.DictType>>(
    '/admin/v1/dict-type',
    {
      params,
    },
  );
}

/**
 * 获取字典类型
 */
async function getDictTypeApi(id: string) {
  return requestClient.get<DictionaryDictTypeApi.DictType>(
    `/admin/v1/dict-type/${id}`,
  );
}

/**
 * 根据字典编码获取字典类型及字典数据（供业务页面消费）
 */
async function getDictTypeByCodeApi(code: string) {
  return requestClient.get<DictionaryDictTypeApi.DictType>(
    `/admin/v1/dict-type/${code}/data`,
  );
}

/**
 * 创建字典类型
 */
async function createDictTypeApi(
  data: Omit<DictionaryDictTypeApi.DictType, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/dict-type', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新字典类型
 */
async function updateDictTypeApi(
  id: string,
  data: Omit<DictionaryDictTypeApi.DictType, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/dict-type/${id}`, data, {
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
  createDictTypeApi,
  deleteDictTypeApi,
  getDictTypeApi,
  getDictTypeByCodeApi,
  getDictTypeListApi,
  updateDictTypeApi,
  updateDictTypeStatusApi,
};
