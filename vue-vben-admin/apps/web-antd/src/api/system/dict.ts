import { requestClient } from '#/api/request';

export namespace SystemDictApi {
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
    /** 描述 */
    description?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  /** 字典数据 */
  export interface DictData {
    /** 数据ID */
    id: string;
    /** 所属字典类型ID */
    dictTypeId: string;
    /** 显示标签 */
    label: string;
    /** 实际值 */
    value: string;
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
 * 获取字典类型列表
 */
async function getDictTypeListApi() {
  return requestClient.get<Array<SystemDictApi.DictType>>('/admin/v1/dict-type');
}

/**
 * 创建字典类型
 */
async function createDictTypeApi(
  data: Omit<SystemDictApi.DictType, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/dict-type', data);
}

/**
 * 更新字典类型
 */
async function updateDictTypeApi(
  id: string,
  data: Omit<SystemDictApi.DictType, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/dict-type/${id}`, data);
}

/**
 * 更新字典类型状态
 */
async function updateDictTypeStatusApi(
  id: string,
  status: SystemDictApi.DictType['status'],
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

/**
 * 根据字典类型获取字典数据列表
 */
async function getDictDataListByTypeApi(dictTypeId: string) {
  return requestClient.get<Array<SystemDictApi.DictData>>(
    '/admin/v1/dict/data',
    {
      params: { dictTypeId },
    },
  );
}

/**
 * 创建字典数据
 */
async function createDictDataApi(
  data: Omit<SystemDictApi.DictData, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/dict/data', data);
}

/**
 * 更新字典数据
 */
async function updateDictDataApi(
  id: string,
  data: Omit<SystemDictApi.DictData, 'id' | 'createdAt' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/dict/data/${id}`, data);
}

/**
 * 更新字典数据状态
 */
async function updateDictDataStatusApi(
  id: string,
  status: SystemDictApi.DictData['status'],
) {
  return requestClient.put(`/admin/v1/dict/data/${id}/status`, {
    status,
  });
}

/**
 * 删除字典数据
 */
async function deleteDictDataApi(ids: string[]) {
  return requestClient.delete('/admin/v1/dict/data', {
    params: {
      ids,
    },
  });
}

export {
  createDictDataApi,
  createDictTypeApi,
  deleteDictDataApi,
  deleteDictTypeApi,
  getDictDataListByTypeApi,
  getDictTypeListApi,
  updateDictDataApi,
  updateDictDataStatusApi,
  updateDictTypeApi,
  updateDictTypeStatusApi,
};

