import type { AxiosRequestConfig } from "axios";

export interface RequestOptions {
  /** 是否显示错误提示 */
  showErrorMessage?: boolean;

  /** 是否允许重复请求 */
  allowRepeatRequest?: boolean;

  /** 是否返回原始 response */
  returnRawResponse?: boolean;

  /** 自定义成功状态码 */
  successStatus?: number[];
}

export interface HttpRequestConfig<T = any> extends AxiosRequestConfig<T> {
  requestOptions?: RequestOptions;
}

export interface HttpError {
  status: number;
  message: string;
  data?: any;
}
