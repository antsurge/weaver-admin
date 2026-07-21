import { requestClient } from '#/api/request';
import type { AllResult } from '#/types/pagination'

export namespace OrganizationDepartmentApi {
  /** 部门 */
  export interface Department {
    /** ID */
    id: string;
    /** 父部门ID，空表示根节点 */
    parentID: string;
    /** 名称 */
    name: string;
    /** 编码 */
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
    /** 子级部门 */
    children?: Department[];
  }

  export interface DepartmentTreeParams {
    name?: string;
    code?: string;
    status?: 'enabled' | 'disabled';
  }
}

/**
 * 获取部门树
 */
async function getDepartmentTreeApi(params?: OrganizationDepartmentApi.DepartmentTreeParams) {
  return requestClient.get<AllResult<OrganizationDepartmentApi.Department>>(
    '/admin/v1/department/tree',
    {
      params: params
    }
  );
}

/**
 * 获取部门
 */
async function getDepartmentApi(id:string) {
  return requestClient.get<OrganizationDepartmentApi.Department>(
    `/admin/v1/department/${id}`,
  );
}

/**
 * 创建部门
 * @param data 部门数据
 */
async function createDepartmentApi(
  data: Omit<OrganizationDepartmentApi.Department, 'id' | 'children'>,
) {
  return requestClient.post('/admin/v1/department', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新部门
 *
 * @param id 部门 ID
 * @param data 部门数据
 */
async function updateDepartmentApi(
  id: string,
  data: Omit<OrganizationDepartmentApi.Department, 'id' | 'children'>,
) {
  return requestClient.put(`/admin/v1/department/${id}`, data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新部门状态
 *
 * @param id 部门 ID
 * @param status 状态
 */
async function updateDepartmentStatusApi(
  id: string,
  status: OrganizationDepartmentApi.Department['status'],
) {
  return requestClient.put(`/admin/v1/department/${id}/status`, {
    status,
  });
}

/**
 * 删除部门
 * @param ids 部门 ID 集合
 */
async function deleteDepartmentApi(ids: string[]) {
  return requestClient.delete('/admin/v1/department', {
    params: {
      ids,
    },
  });
}

export {
  getDepartmentTreeApi,
  getDepartmentApi,
  createDepartmentApi,
  updateDepartmentApi,
  updateDepartmentStatusApi,
  deleteDepartmentApi,
};

