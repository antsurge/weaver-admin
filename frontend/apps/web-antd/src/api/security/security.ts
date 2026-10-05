import { requestClient } from '#/api/request';

export namespace SecurityApi {
  /** IP 访问控制模式 */
  export type IPMode = 'blacklist' | 'off' | 'whitelist';

  /** 安全策略（登录策略 + 访问控制） */
  export interface SecurityPolicy {
    /** 是否启用登录失败次数限制 */
    loginFailEnabled: boolean;
    /** 统计窗口内允许的最大失败次数 */
    loginFailMax: number;
    /** 失败次数统计窗口（分钟） */
    loginFailWindowMinutes: number;
    /** 达到上限后的锁定时长（分钟） */
    loginLockMinutes: number;
    /** IP 访问控制模式 */
    ipMode: IPMode;
    /** IP 列表（支持 CIDR） */
    ipList: string[];
    /** access token 有效期（秒），范围 60-2592000，0 回退默认值 7200 */
    accessTokenTtlSeconds: number;
    /** refresh token 有效期（秒），范围 300-31536000，0 回退默认值 604800 */
    refreshTokenTtlSeconds: number;
    /** 备注 */
    remark?: string;
    /** 更新时间 */
    updatedAt?: string;
  }
}

/** protojson 会把 int64 序列化为字符串，这里统一归一化为 number */
function toNumber(v: unknown): number {
  const n = Number(v ?? 0);
  return Number.isFinite(n) ? n : 0;
}

function normalize(
  raw: SecurityApi.SecurityPolicy,
): SecurityApi.SecurityPolicy {
  return {
    ...raw,
    loginFailMax: toNumber(raw.loginFailMax),
    loginFailWindowMinutes: toNumber(raw.loginFailWindowMinutes),
    loginLockMinutes: toNumber(raw.loginLockMinutes),
    accessTokenTtlSeconds: toNumber(raw.accessTokenTtlSeconds),
    refreshTokenTtlSeconds: toNumber(raw.refreshTokenTtlSeconds),
    ipList: Array.isArray(raw.ipList) ? raw.ipList : [],
  };
}

/**
 * 获取安全策略
 */
async function getSecurityPolicyApi() {
  const res = await requestClient.get<SecurityApi.SecurityPolicy>(
    '/admin/v1/security/policy',
  );
  return normalize(res);
}

/**
 * 更新安全策略（保存后即时生效）
 */
async function updateSecurityPolicyApi(
  data: SecurityApi.SecurityPolicy,
): Promise<SecurityApi.SecurityPolicy> {
  const res = await requestClient.put<SecurityApi.SecurityPolicy>(
    '/admin/v1/security/policy',
    data,
    { showSuccessMessage: true },
  );
  return normalize(res);
}

export { getSecurityPolicyApi, updateSecurityPolicyApi };
