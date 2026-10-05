import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace FormSchemaApi {
  /** 表单定义 */
  export interface FormSchema {
    /** 表单ID */
    id: string;
    /** 表单名称 */
    name: string;
    /** 表单编码（唯一，运行时渲染时引用） */
    code: string;
    /** 表单描述 */
    description?: string;
    /** 表单 JSON Schema 定义 */
    schemaJson?: Record<string, unknown>;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'disabled' | 'enabled';
    /** 备注 */
    remark?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  export interface FormSchemaListParams extends PaginationParams {
    name?: string;
    code?: string;
    status?: string;
  }

  /** 表单提交记录 */
  export interface FormSubmission {
    /** 提交记录ID */
    id: string;
    /** 表单编码 */
    formCode: string;
    /** 表单名称 */
    formName?: string;
    /** 提交数据（字段 key -> 值） */
    submitData?: Record<string, unknown>;
    /** 提交时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }
}

/**
 * 获取表单定义列表
 */
async function getFormSchemaListApi(
  params?: FormSchemaApi.FormSchemaListParams,
) {
  return requestClient.get<PaginationResult<FormSchemaApi.FormSchema>>(
    '/admin/v1/form-schemas',
    { params },
  );
}

/**
 * 根据编码获取表单定义（运行时渲染）
 */
async function getFormSchemaByCodeApi(code: string) {
  return requestClient.get<FormSchemaApi.FormSchema>(
    `/admin/v1/form-schemas/${code}/detail`,
  );
}

/**
 * 创建表单定义
 */
async function createFormSchemaApi(
  data: Omit<FormSchemaApi.FormSchema, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/form-schemas', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新表单定义
 */
async function updateFormSchemaApi(
  id: string,
  data: Omit<FormSchemaApi.FormSchema, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/form-schemas/${id}`, data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新表单状态
 */
async function updateFormSchemaStatusApi(
  id: string,
  status: FormSchemaApi.FormSchema['status'],
) {
  return requestClient.put(`/admin/v1/form-schemas/${id}/status`, { status });
}

/**
 * 删除表单定义（批量）
 */
async function deleteFormSchemaApi(ids: string[]) {
  return requestClient.delete('/admin/v1/form-schemas', {
    params: { ids },
  });
}

/**
 * 提交表单数据（运行时渲染表单的用户提交）
 */
async function submitFormDataApi(code: string, data: Record<string, unknown>) {
  return requestClient.post<{ id: string }>(
    `/admin/v1/form-schemas/${code}/submit`,
    { data },
    { showSuccessMessage: true },
  );
}

/**
 * 查询表单提交记录（分页）
 */
async function getFormSubmissionListApi(
  code: string,
  params?: PaginationParams,
) {
  return requestClient.get<PaginationResult<FormSchemaApi.FormSubmission>>(
    `/admin/v1/form-schemas/${code}/submissions`,
    { params },
  );
}

export {
  createFormSchemaApi,
  deleteFormSchemaApi,
  getFormSchemaByCodeApi,
  getFormSchemaListApi,
  getFormSubmissionListApi,
  submitFormDataApi,
  updateFormSchemaApi,
  updateFormSchemaStatusApi,
};
