import type { PaginationParams, PaginationResult } from '#/types/pagination';

import { useAppConfig } from '@vben/hooks';
import { useAccessStore } from '@vben/stores';

import { requestClient } from '#/api/request';

export namespace NotificationApi {
  /** 通知正文 */
  export interface Notification {
    id: string;
    title: string;
    content: string;
    /** system=系统 announce=公告 audit=审批 todo=待办 alert=告警 */
    type: string;
    /** info=普通 success=成功 warn=警告 error=严重 */
    level: string;
    bizType?: string;
    bizId?: string;
    senderId?: string;
    senderName?: string;
    /** 点击跳转的前端路由 */
    link?: string;
    /** all=全体 user=指定用户 role=指定角色 */
    targetType: string;
    targetIds?: string[];
    status: string;
    publishedAt?: string;
    expireAt?: string;
    createdAt?: string;
    updatedAt?: string;
  }

  /** 用户收件箱条目 */
  export interface NotificationRecord {
    id: string;
    notificationId: string;
    userId: string;
    isRead: boolean;
    readAt?: string;
    isDeleted: boolean;
    createdAt?: string;
    updatedAt?: string;
    /** 通知正文（列表接口联带返回） */
    notification?: Notification;
  }

  export interface ListParams extends PaginationParams {
    /** 是否已读（不传=全部） */
    isRead?: boolean;
    type?: string;
    level?: string;
    keyword?: string;
  }

  export interface ManageListParams extends PaginationParams {
    title?: string;
    type?: string;
    level?: string;
    status?: string;
  }

  export interface CreateParams {
    title: string;
    content: string;
    type?: string;
    level?: string;
    bizType?: string;
    bizId?: string;
    link?: string;
    targetType?: 'all' | 'role' | 'user';
    targetIds?: string[];
  }

  /** SSE 推送事件（与后端 biz.NotificationEvent 对应） */
  export interface NotificationEvent {
    /** created=新通知 revoked=撤回 */
    action: 'created' | 'revoked';
    id: string;
    title: string;
    content: string;
    type: string;
    level: string;
    link: string;
    createdAt: string;
  }
}

/** 我的消息列表 */
async function getNotificationListApi(params?: NotificationApi.ListParams) {
  return requestClient.get<
    PaginationResult<NotificationApi.NotificationRecord>
  >('/admin/v1/notifications', { params });
}

/** 未读数量 */
async function getUnreadCountApi() {
  return requestClient.get<{ count: number }>(
    '/admin/v1/notifications/unread-count',
  );
}

/** 标记已读 */
async function readNotificationApi(id: string) {
  return requestClient.put(`/admin/v1/notifications/${id}/read`, { id });
}

/** 全部已读 */
async function readAllNotificationApi() {
  return requestClient.put('/admin/v1/notifications/read-all', {});
}

/** 删除消息（支持批量） */
async function deleteNotificationApi(ids: string[]) {
  return requestClient.delete('/admin/v1/notifications', {
    params: { ids },
  });
}

/** 管理端：通知列表 */
async function getNotificationManageListApi(
  params?: NotificationApi.ManageListParams,
) {
  return requestClient.get<PaginationResult<NotificationApi.Notification>>(
    '/admin/v1/notifications/manage',
    { params },
  );
}

/** 管理端：发布通知 */
async function createNotificationApi(data: NotificationApi.CreateParams) {
  return requestClient.post<NotificationApi.Notification>(
    '/admin/v1/notifications',
    data,
  );
}

/** 管理端：撤回通知 */
async function revokeNotificationApi(id: string) {
  return requestClient.put(`/admin/v1/notifications/${id}/revoke`, { id });
}

const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);

/**
 * 订阅实时通知推送（SSE）
 *
 * EventSource 不支持自定义 Header，token 通过 query 参数传递
 * （后端 pkg/middleware/auth 已支持 ?token=）。
 *
 * @returns 取消订阅函数
 */
function subscribeNotificationStream(handlers: {
  onError?: () => void;
  onNotification?: (ev: NotificationApi.NotificationEvent) => void;
  onOpen?: () => void;
}): () => void {
  const accessStore = useAccessStore();
  const token = accessStore.accessToken;
  if (!token) {
    return () => {};
  }

  const url = `${apiURL}/admin/v1/notifications/stream?token=${encodeURIComponent(token)}`;
  const source = new EventSource(url, { withCredentials: false });

  source.addEventListener('connected', () => handlers.onOpen?.());
  source.addEventListener('notification', (e) => {
    try {
      const ev = JSON.parse((e as MessageEvent).data);
      handlers.onNotification?.(ev);
    } catch {
      // 忽略无法解析的事件帧
    }
  });
  source.addEventListener('error', () => handlers.onError?.());

  return () => source.close();
}

export {
  createNotificationApi,
  deleteNotificationApi,
  getNotificationListApi,
  getNotificationManageListApi,
  getUnreadCountApi,
  readAllNotificationApi,
  readNotificationApi,
  revokeNotificationApi,
  subscribeNotificationStream,
};
