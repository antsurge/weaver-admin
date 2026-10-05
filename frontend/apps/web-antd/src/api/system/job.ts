import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace SystemJobApi {
  /** 定时任务 */
  export interface Job {
    /** 任务ID */
    id: string;
    /** 任务名称 */
    name: string;
    /** 任务分组 */
    jobGroup?: string;
    /** 任务类型 */
    jobType?: string;
    /** 调用目标（HTTP URL） */
    invokeTarget: string;
    /** cron 表达式 */
    cronExpression: string;
    /** 错失执行策略 */
    misfirePolicy?: string;
    /** 是否允许并发执行 */
    concurrent?: boolean;
    /** 状态：enabled=启用 disabled=禁用 */
    status: 'disabled' | 'enabled';
    /** 备注 */
    remark?: string;
    /** 下次执行时间 */
    nextRunTime?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  /** 任务执行日志 */
  export interface JobLog {
    /** 日志ID */
    id: string;
    /** 任务ID */
    jobId: string;
    /** 任务名称 */
    jobName: string;
    /** 任务分组 */
    jobGroup?: string;
    /** 调用目标 */
    invokeTarget?: string;
    /** cron 表达式 */
    cronExpression?: string;
    /** 状态：success=成功 fail=失败 */
    status: 'fail' | 'success';
    /** 执行信息 */
    jobMessage?: string;
    /** 开始时间 */
    startTime?: string;
    /** 结束时间 */
    endTime?: string;
    /** 耗时（毫秒） */
    duration?: number;
    /** 创建时间 */
    createdAt?: string;
  }

  export interface JobListParams extends PaginationParams {
    /** 任务名称（模糊） */
    name?: string;
    /** 任务分组（模糊） */
    jobGroup?: string;
    /** 状态 */
    status?: string;
  }

  export interface JobLogListParams extends PaginationParams {
    /** 任务ID */
    jobId?: string;
    /** 任务名称（模糊） */
    jobName?: string;
    /** 状态 */
    status?: string;
  }
}

/**
 * 获取任务列表
 */
async function getJobListApi(params?: SystemJobApi.JobListParams) {
  return requestClient.get<PaginationResult<SystemJobApi.Job>>(
    '/admin/v1/jobs',
    { params },
  );
}

/**
 * 获取任务详情
 */
async function getJobApi(id: string) {
  return requestClient.get<SystemJobApi.Job>(`/admin/v1/jobs/${id}`);
}

/**
 * 创建任务
 */
async function createJobApi(
  data: Omit<SystemJobApi.Job, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/jobs', data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新任务
 */
async function updateJobApi(
  id: string,
  data: Omit<SystemJobApi.Job, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.put(`/admin/v1/jobs/${id}`, data, {
    showSuccessMessage: true,
  });
}

/**
 * 更新任务状态
 */
async function updateJobStatusApi(
  id: string,
  status: SystemJobApi.Job['status'],
) {
  return requestClient.put(`/admin/v1/jobs/${id}/status`, { status });
}

/**
 * 立即执行一次
 */
async function runJobApi(id: string) {
  return requestClient.post(`/admin/v1/jobs/${id}/run`, {}, {});
}

/**
 * 删除任务（批量）
 */
async function deleteJobApi(ids: string[]) {
  return requestClient.delete('/admin/v1/jobs', {
    params: { ids },
  });
}

/**
 * 获取任务执行日志列表
 */
async function getJobLogListApi(params?: SystemJobApi.JobLogListParams) {
  return requestClient.get<PaginationResult<SystemJobApi.JobLog>>(
    '/admin/v1/job-logs',
    { params },
  );
}

/**
 * 获取任务执行日志详情
 */
async function getJobLogApi(id: string) {
  return requestClient.get<SystemJobApi.JobLog>(`/admin/v1/job-logs/${id}`);
}

/**
 * 删除任务日志（批量）
 */
async function deleteJobLogApi(ids: string[]) {
  return requestClient.delete('/admin/v1/job-logs', {
    params: { ids },
  });
}

/**
 * 清空任务日志
 */
async function clearJobLogApi() {
  return requestClient.delete('/admin/v1/job-logs/clear');
}

export {
  clearJobLogApi,
  createJobApi,
  deleteJobApi,
  deleteJobLogApi,
  getJobApi,
  getJobListApi,
  getJobLogApi,
  getJobLogListApi,
  runJobApi,
  updateJobApi,
  updateJobStatusApi,
};
