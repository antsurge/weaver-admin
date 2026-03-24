import { requestClient } from '#/api/request';
import type { PaginationParams,AllResult } from '#/types/pagination'

export namespace SystemOrganizationApi {
  /** 组织机构 */
  export interface Organization {
    /** 部门ID */
    id: string;
    /** 父部门ID，空表示根节点 */
    parentId: string;
    /** 部门名称 */
    name: string;
    /** 部门编码 */
    code?: string;
    /** 权重 */
    weight: number;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'enabled' | 'disabled';
    /** 负责人姓名 */
    leaderName?: string;
    /** 联系电话 */
    leaderPhone?: string;
    /** 邮箱 */
    leaderEmail?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
    /** 子级组织 */
    children?: Organization[];
  }

  export interface OrganizationTreeParams{
    name?: string;
    code?: string;
    status?: 'enabled' | 'disabled';
  }
}

/**
 * 获取组织机构树
 */
async function getOrganizationTreeApi(params?:SystemOrganizationApi.OrganizationTreeParams) {
  return requestClient.get<AllResult<SystemOrganizationApi.Organization>>(
    '/admin/v1/organization/tree',
    {
      params:params
    }
  );
}

/**
 * 创建组织机构
 * @param data 组织数据
 */
async function createOrganization(
  data: Omit<SystemOrganizationApi.Organization, 'id' | 'children'>,
) {
  return requestClient.post('/admin/v1/organization', data);
}

/**
 * 更新组织机构
 *
 * @param id 组织 ID
 * @param data 组织数据
 */
async function updateOrganization(
  id: string,
  data: Omit<SystemOrganizationApi.Organization, 'id' | 'children'>,
) {
  return requestClient.put(`/admin/v1/organization/${id}`, data);
}

/**
 * 更新组织机构状态
 *
 * @param id 组织 ID
 * @param status 状态
 */
async function updateOrganizationStatusApi(
  id: string,
  status: SystemOrganizationApi.Organization['status'],
) {
  return requestClient.put(`/admin/v1/organization/${id}/status`, {
    status,
  });
}

/**
 * 删除组织机构
 * @param ids 组织 ID 集合
 */
async function deleteOrganizationApi(ids: string[]) {
  return requestClient.delete('/admin/v1/organization', {
    params: {
      ids,
    },
  });
}

export {
  createOrganization,
  deleteOrganizationApi,
  getOrganizationTreeApi,
  updateOrganization,
  updateOrganizationStatusApi,
};

