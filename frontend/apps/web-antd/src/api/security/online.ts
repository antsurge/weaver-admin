import { requestClient } from '#/api/request';

export namespace OnlineApi {
  /** 在线用户（按用户聚合） */
  export interface OnlineUser {
    /** 用户ID */
    id: string;
    /** 用户名 */
    username?: string;
    /** 姓名 */
    realName?: string;
    /** 最近登录 IP */
    ip?: string;
    /** 浏览器/设备 UA */
    userAgent?: string;
    /** 最近登录时间（Unix 秒） */
    loginAt?: number;
    /** 会话过期时间（Unix 秒） */
    expireAt?: number;
    /** 在线会话数（多设备 >1） */
    sessionCount?: number;
  }
}

/** 在线用户列表 */
async function getOnlineUsersApi() {
  return requestClient.get<{
    items: OnlineApi.OnlineUser[];
    total: number;
  }>('/admin/v1/online-users');
}

/** 强制下线指定用户（踢掉其全部在线会话） */
async function forceLogoutApi(id: string) {
  return requestClient.post(`/admin/v1/online-users/${id}/logout`, { id });
}

export { forceLogoutApi, getOnlineUsersApi };
