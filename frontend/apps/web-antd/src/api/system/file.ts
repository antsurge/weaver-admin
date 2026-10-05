import { useAppConfig } from '@vben/hooks';

import { requestClient } from '#/api/request';

/** 与请求客户端一致的 API 基础地址（如 /api），用于拼接可被 dev 代理转发的图片地址 */
const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);
const apiBase = (apiURL || '').replace(/\/$/, '');

export namespace SystemFileApi {
  /** 上传结果 */
  export interface UploadResult {
    /** 对象存储中的对象键（删除时使用，也是数据库应存储的值） */
    objectKey: string;
    /** 文件访问地址（上传即时返回，私有桶下可能带签名，勿持久化） */
    url: string;
    /** 原始文件名 */
    fileName: string;
    /** 文件大小（字节） */
    size: number;
  }

  /** 上传参数 */
  export interface UploadParams {
    /** 文件 */
    file: Blob | File;
    /** 自定义存储目录（可选） */
    dir?: string;
  }
}

/**
 * 根据对象键生成可访问的文件地址。
 * 后端会针对私有桶动态生成带签名的临时地址并重定向，因此该地址可安全地
 * 直接用于 <img src>，无需携带认证头，也不会过期。
 *
 * @param objectKey 对象键，如 uploads/avatar/xxx.png
 */
function getFileAccessUrl(objectKey?: string): string {
  if (!objectKey) return '';
  // 已经是完整 URL（历史数据兼容）时原样返回
  if (/^https?:\/\//i.test(objectKey)) return objectKey;
  const key = objectKey.replace(/^\/+/, '');
  // 使用 query 形式：objectKey 内部含 "/" 需编码，后端从 query 读取后生成预签名地址并 302。
  // 必须带上 apiBase（如 /api）前缀，才能被 dev server 代理转发到后端。
  return `${apiBase}/admin/v1/file?objectKey=${encodeURIComponent(key)}`;
}

/**
 * 上传文件到对象存储
 * 对应后端 POST /admin/v1/file/upload
 */
async function uploadFileApi(
  params: SystemFileApi.UploadParams,
  config?: { showSuccessMessage?: boolean },
) {
  return requestClient.upload<SystemFileApi.UploadResult>(
    '/admin/v1/file/upload',
    { ...params },
    { showSuccessMessage: false, ...config },
  );
}

/** 删除对象存储中的文件 */
async function deleteFileApi(objectKey: string) {
  return requestClient.delete('/admin/v1/file', { params: { objectKey } });
}

export { deleteFileApi, getFileAccessUrl, uploadFileApi };
