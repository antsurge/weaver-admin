import { requestClient } from '#/api/request';

export namespace MonitorApi {
  /** 服务信息 */
  export interface ServerInfo {
    hostname: string;
    os: string;
    arch: string;
    goVersion: string;
    numCPU: number;
    numGoroutine: number;
    goMaxProcs: number;
    heapAlloc: number;
    heapSys: number;
    heapInuse: number;
    stackInuse: number;
    sys: number;
    totalAlloc: number;
    numGC: number;
    gcPauseTotalMs: number;
    startTime?: string;
    uptimeSeconds: number;
  }

  /** 缓存信息 */
  export interface CacheInfo {
    version: string;
    mode: string;
    role: string;
    uptimeSeconds: number;
    connectedClients: number;
    usedMemory: number;
    usedMemoryHuman: string;
    maxMemory: number;
    totalKeys: number;
    keyspaceHits: number;
    keyspaceMisses: number;
    hitRate: number;
    totalCommands: number;
    expiredKeys: number;
    evictedKeys: number;
    opsPerSec: number;
  }

  /** 缓存键 */
  export interface CacheKey {
    key: string;
    rawKey: string;
    type: string;
    ttlSeconds: number;
    size: number;
    value: string;
  }

  /** 表统计 */
  export interface DbTableStat {
    name: string;
    rows: number;
    dataSize: number;
    indexSize: number;
  }

  /** 数据库信息 */
  export interface DatabaseInfo {
    driver: string;
    version: string;
    database: string;
    tableCount: number;
    maxOpenConnections: number;
    openConnections: number;
    inUse: number;
    idle: number;
    waitCount: number;
    waitDurationMs: number;
    maxIdleClosed: number;
    maxLifetimeClosed: number;
    tables: DbTableStat[];
  }

  /** 接口指标 */
  export interface ApiMetric {
    method: string;
    path: string;
    total: number;
    success: number;
    error: number;
    errorRate: number;
    avgMs: number;
    maxMs: number;
    lastMs: number;
  }

  export interface ApiMetricsResult {
    items: ApiMetric[];
    totalRequests: number;
    totalErrors: number;
    uptimeSeconds: number;
  }

  /** 任务执行记录 */
  export interface JobRunLog {
    id: string;
    jobName: string;
    jobGroup: string;
    status: string;
    message: string;
    startTime?: string;
    duration: number;
  }

  /** 定时任务监控 */
  export interface JobMonitorInfo {
    totalJobs: number;
    enabledJobs: number;
    disabledJobs: number;
    totalRuns: number;
    successRuns: number;
    failedRuns: number;
    successRate: number;
    runningJobs: number;
    recentRuns: JobRunLog[];
  }
}

type Raw = Record<string, any>;

/** protojson 会把 int64 序列化为字符串，统一归一化为 number */
function n(v: unknown): number {
  const num = Number(v ?? 0);
  return Number.isFinite(num) ? num : 0;
}

async function getServerInfoApi(): Promise<MonitorApi.ServerInfo> {
  const raw = await requestClient.get<Raw>('/admin/v1/monitor/server');
  return {
    ...raw,
    numCPU: n(raw.numCPU),
    numGoroutine: n(raw.numGoroutine),
    goMaxProcs: n(raw.goMaxProcs),
    heapAlloc: n(raw.heapAlloc),
    heapSys: n(raw.heapSys),
    heapInuse: n(raw.heapInuse),
    stackInuse: n(raw.stackInuse),
    sys: n(raw.sys),
    totalAlloc: n(raw.totalAlloc),
    numGC: n(raw.numGC),
    gcPauseTotalMs: n(raw.gcPauseTotalMs),
    uptimeSeconds: n(raw.uptimeSeconds),
  } as MonitorApi.ServerInfo;
}

async function getCacheInfoApi(): Promise<MonitorApi.CacheInfo> {
  const raw = await requestClient.get<Raw>('/admin/v1/monitor/cache');
  return {
    ...raw,
    uptimeSeconds: n(raw.uptimeSeconds),
    connectedClients: n(raw.connectedClients),
    usedMemory: n(raw.usedMemory),
    maxMemory: n(raw.maxMemory),
    totalKeys: n(raw.totalKeys),
    keyspaceHits: n(raw.keyspaceHits),
    keyspaceMisses: n(raw.keyspaceMisses),
    totalCommands: n(raw.totalCommands),
    expiredKeys: n(raw.expiredKeys),
    evictedKeys: n(raw.evictedKeys),
    opsPerSec: n(raw.opsPerSec),
  } as MonitorApi.CacheInfo;
}

async function getCacheKeysApi(params: {
  currentPage: number;
  pageSize: number;
  pattern?: string;
}): Promise<{ items: MonitorApi.CacheKey[]; total: number }> {
  const raw = await requestClient.get<Raw>('/admin/v1/monitor/cache/keys', {
    params,
  });
  return {
    total: n(raw.total),
    items: (raw.items ?? []).map((i: Raw) => ({
      key: i.key,
      rawKey: i.rawKey,
      type: i.type,
      ttlSeconds: n(i.ttlSeconds),
      size: n(i.size),
      value: i.value,
    })),
  };
}

async function getDatabaseInfoApi(): Promise<MonitorApi.DatabaseInfo> {
  const raw = await requestClient.get<Raw>('/admin/v1/monitor/database');
  return {
    ...raw,
    tableCount: n(raw.tableCount),
    maxOpenConnections: n(raw.maxOpenConnections),
    openConnections: n(raw.openConnections),
    inUse: n(raw.inUse),
    idle: n(raw.idle),
    waitCount: n(raw.waitCount),
    waitDurationMs: n(raw.waitDurationMs),
    maxIdleClosed: n(raw.maxIdleClosed),
    maxLifetimeClosed: n(raw.maxLifetimeClosed),
    tables: (raw.tables ?? []).map((t: Raw) => ({
      name: t.name,
      rows: n(t.rows),
      dataSize: n(t.dataSize),
      indexSize: n(t.indexSize),
    })),
  } as MonitorApi.DatabaseInfo;
}

async function getApiMetricsApi(params?: {
  method?: string;
  path?: string;
}): Promise<MonitorApi.ApiMetricsResult> {
  const raw = await requestClient.get<Raw>('/admin/v1/monitor/metrics', {
    params,
  });
  return {
    totalRequests: n(raw.totalRequests),
    totalErrors: n(raw.totalErrors),
    uptimeSeconds: n(raw.uptimeSeconds),
    items: (raw.items ?? []).map((i: Raw) => ({
      ...i,
      total: n(i.total),
      success: n(i.success),
      error: n(i.error),
    })),
  };
}

async function getJobMonitorApi(params?: {
  group?: string;
  name?: string;
}): Promise<MonitorApi.JobMonitorInfo> {
  const raw = await requestClient.get<Raw>('/admin/v1/monitor/job', { params });
  return {
    ...raw,
    totalJobs: n(raw.totalJobs),
    enabledJobs: n(raw.enabledJobs),
    disabledJobs: n(raw.disabledJobs),
    totalRuns: n(raw.totalRuns),
    successRuns: n(raw.successRuns),
    failedRuns: n(raw.failedRuns),
    runningJobs: n(raw.runningJobs),
    recentRuns: (raw.recentRuns ?? []).map((r: Raw) => ({
      ...r,
      duration: n(r.duration),
    })),
  } as MonitorApi.JobMonitorInfo;
}

export {
  getApiMetricsApi,
  getCacheInfoApi,
  getCacheKeysApi,
  getDatabaseInfoApi,
  getJobMonitorApi,
  getServerInfoApi,
};
