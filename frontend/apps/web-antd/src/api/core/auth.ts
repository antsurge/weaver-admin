import type { RequestResponse } from '@vben/request';

import { useAccessStore } from '@vben/stores';

import { baseRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    captcha?: string;
    captchaId?: string;
    password?: string;
    username?: string;
    /** 租户编码（multi 租户模式下必填；single 模式可省略） */
    tenantCode?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    accessToken: string;
    accessTokenExpiresIn: number;
    refreshToken: string;
    refreshTokenExpiresIn: number;
    tokenType: string;
    userId: string;
  }
}

/**
 * refreshToken 本地存储
 * 后端刷新接口要求同时携带 refreshToken/accessToken，二者需成对保存
 */
const REFRESH_TOKEN_KEY = 'weaver-admin:refresh-token';

function getRefreshToken(): null | string {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

function setRefreshToken(token: null | string) {
  if (token) {
    localStorage.setItem(REFRESH_TOKEN_KEY, token);
  } else {
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  }
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  return requestClient.post<AuthApi.LoginResult>('/admin/v1/login', data);
}

/**
 * 获取登录图形验证码
 * 后端: GET /admin/v1/get-captcha → { captchaId, imageBase64 }
 */
export async function getCaptchaApi() {
  return requestClient.get<{ captchaId: string; imageBase64: string }>(
    '/admin/v1/get-captcha',
  );
}

/**
 * 刷新accessToken
 * 后端: POST /admin/v1/refresh-token，body 需同时携带 refreshToken 与 accessToken
 * 使用不带拦截器的 baseRequestClient，避免 401 时触发刷新死循环
 */
async function refreshTokenApi() {
  const accessStore = useAccessStore();

  return baseRequestClient.post<RequestResponse<AuthApi.LoginResult>>(
    '/admin/v1/refresh-token',
    {
      accessToken: accessStore.accessToken ?? '',
      refreshToken: getRefreshToken() ?? '',
    },
  );
}

/**
 * 退出登录（同时作废服务端的 access/refresh token）
 */
async function logoutApi() {
  const accessStore = useAccessStore();

  return requestClient.post('/admin/v1/logout', {
    accessToken: accessStore.accessToken ?? '',
    refreshToken: getRefreshToken() ?? '',
  });
}

export { getRefreshToken, logoutApi, refreshTokenApi, setRefreshToken };
