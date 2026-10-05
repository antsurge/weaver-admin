/**
 * 请求客户端
 * 适配 RESTful API:
 * 成功: HTTP 200
 * 失败: HTTP 4xx / 5xx
 */

import type { RequestClientOptions } from '@vben/request';

import { useAppConfig } from '@vben/hooks';
import { preferences } from '@vben/preferences';
import {
  authenticateResponseInterceptor,
  errorMessageResponseInterceptor,
  RequestClient,
} from '@vben/request';
import { useAccessStore } from '@vben/stores';

import { message } from 'ant-design-vue';

import { useAuthStore } from '#/store';

import { refreshTokenApi, setRefreshToken } from './core';

const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);

/**
 * token格式
 */
function formatToken(token: null | string) {
  return token ? `Bearer ${token}` : null;
}

/**
 * 创建请求客户端
 */
function createRequestClient(baseURL: string, options?: RequestClientOptions) {
  const client = new RequestClient({
    ...options,
    baseURL,
  });

  /**
   * token失效处理
   */
  async function doReAuthenticate() {
    console.warn('Token expired, need re-login.');

    const accessStore = useAccessStore();
    const authStore = useAuthStore();

    accessStore.setAccessToken(null);
    setRefreshToken(null);

    if (
      preferences.app.loginExpiredMode === 'modal' &&
      accessStore.isAccessChecked
    ) {
      accessStore.setLoginExpired(true);
    } else {
      await authStore.logout();
    }
  }

  /**
   * 刷新token
   * 后端返回完整 LoginResponse，需同时更新 accessToken 与 refreshToken
   */
  async function doRefreshToken() {
    const accessStore = useAccessStore();

    // baseRequestClient 为 raw 响应，resp.data 即响应体（LoginResponse）
    const resp = await refreshTokenApi();

    const newAuth = resp.data;

    accessStore.setAccessToken(newAuth?.accessToken);
    setRefreshToken(newAuth?.refreshToken);

    return newAuth?.accessToken;
  }

  /**
   * 请求拦截
   */
  client.addRequestInterceptor({
    fulfilled: async (config) => {
      const accessStore = useAccessStore();

      // token
      config.headers.Authorization = formatToken(accessStore.accessToken);

      // 语言
      config.headers['Accept-Language'] = preferences.app.locale;

      return config;
    },
  });

  /**
   * token过期处理
   */
  client.addResponseInterceptor(
    authenticateResponseInterceptor({
      client,
      doReAuthenticate,
      doRefreshToken,
      enableRefreshToken: preferences.app.enableRefreshToken,
      formatToken,
    }),
  );

  /**
   * 成功提示拦截器（🔥新增）
   */
  client.addResponseInterceptor({
    fulfilled: (response) => {
      const config = response.config;
      const resData = response.data ?? {};

      const shouldShowSuccess =
        config?.showSuccessMessage === true || Boolean(config?.successMessage);

      if (shouldShowSuccess) {
        const msg =
          typeof config?.successMessage === 'string'
            ? config.successMessage
            : resData?.message || '操作成功';

        message.success(msg);
      }

      return response;
    },
  });

  /**
   * 错误响应拦截器
   */
  client.addResponseInterceptor(
    errorMessageResponseInterceptor((msg: string, error) => {
      const config = error?.config;

      if (config?.showFailMessage === false) return;

      const responseData = error?.response?.data ?? {};
      const errorMessage =
        config?.failMessage ||
        responseData?.message ||
        responseData?.error ||
        msg ||
        '请求失败';

      message.error(errorMessage);
    }),
  );

  return client;
}

/**
 * 主请求客户端
 */
export const requestClient = createRequestClient(apiURL, {
  responseReturn: 'data',
});

/**
 * 基础请求客户端 (不带拦截器)
 */
export const baseRequestClient = new RequestClient({
  baseURL: apiURL,
});
