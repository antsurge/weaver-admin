import { axiosInstance,defaultRequestOptions } from './constant'
import { addPending, removePending } from './cancle'
import type { AxiosResponse } from 'axios'
import type { HttpRequestConfig } from './interface'
import { StatusCodes } from 'http-status-codes'

class HttpClient {
  constructor() {
    this.setupInterceptors()
  }

  private setupInterceptors() {
    axiosInstance.interceptors.request.use((config: HttpRequestConfig) => {
      const options = {
        ...defaultRequestOptions,
        ...config.requestOptions,
      }

      config.requestOptions = options

      if (!options.allowRepeatRequest) {
        removePending(config)
        addPending(config)
      }

      const token = localStorage.getItem('token')

      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }

      return config
    })

    axiosInstance.interceptors.response.use(
      (response: AxiosResponse) => {
        const config = response.config as HttpRequestConfig

        removePending(config)

        const options = config.requestOptions ?? defaultRequestOptions

        const { status } = response

        if (!options.successStatus?.includes(status)) {
          const error = {
            status,
            message: response.statusText,
            data: response.data,
          }

          return Promise.reject(error)
        }

        if (options.returnRawResponse) {
          return response
        }

        return response.data
      },
      (error) => {
        if (error.response) {
          const { status, data } = error.response

          const errorInfo = {
            status,
            message: data?.message || error.message,
            data,
          }

          return Promise.reject(errorInfo)
        }

        return Promise.reject(error)
      },
    )
  }

  request<T = any>(config: HttpRequestConfig): Promise<T> {
    return axiosInstance.request<any, T>(config)
  }

  get<T = any>(url: string, params?: any, options?: any) {
    return this.request<T>({
      url,
      method: 'GET',
      params,
      requestOptions: options,
    })
  }

  post<T = any>(url: string, data?: any, options?: any) {
    return this.request<T>({
      url,
      method: 'POST',
      data,
      requestOptions: options,
    })
  }

  put<T = any>(url: string, data?: any, options?: any) {
    return this.request<T>({
      url,
      method: 'PUT',
      data,
      requestOptions: options,
    })
  }

  delete<T = any>(url: string, params?: any, options?: any) {
    return this.request<T>({
      url,
      method: 'DELETE',
      params,
      requestOptions: options,
    })
  }

  download(url: string, params?: any) {
    return axiosInstance.get(url, {
      params,
      responseType: 'blob',
    })
  }
}

export const http = new HttpClient()