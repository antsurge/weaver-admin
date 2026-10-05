<script lang="ts" setup>
import type { NotificationItem } from '@vben/layouts';

import type { NotificationApi } from '#/api/system/notification';

import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { AuthenticationLoginExpiredModal } from '@vben/common-ui';
import { useWatermark } from '@vben/hooks';
import { BookOpenText, CircleHelp, SvgGithubIcon } from '@vben/icons';
import {
  BasicLayout,
  LockScreen,
  Notification,
  UserDropdown,
} from '@vben/layouts';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { openWindow } from '@vben/utils';

import { getFileAccessUrl } from '#/api/system/file';
import {
  deleteNotificationApi,
  getNotificationListApi,
  getUnreadCountApi,
  readAllNotificationApi,
  readNotificationApi,
  subscribeNotificationStream,
} from '#/api/system/notification';
import { $t } from '#/locales';
import { useAuthStore } from '#/store';
import LoginForm from '#/views/_core/authentication/login.vue';
import ProfileModal from '#/views/_core/profile/profile-modal.vue';

const notifications = ref<NotificationItem[]>([]);
const unreadCount = ref(0);

/** 收件记录 → 铃铛展示项 */
function toNotificationItem(
  record: NotificationApi.NotificationRecord,
): NotificationItem {
  const n = record.notification;
  return {
    id: record.id,
    avatar: preferences.app.defaultAvatar,
    date: formatRelativeTime(record.createdAt),
    isRead: record.isRead,
    message: n?.content ?? '',
    title: n?.title ?? '',
    link: n?.link || undefined,
  };
}

/** 时间戳转相对时间，空值返回空串 */
function formatRelativeTime(value?: string) {
  if (!value) return '';
  const ts = new Date(value).getTime();
  if (Number.isNaN(ts)) return '';
  const diff = Date.now() - ts;
  if (diff < 60_000) return '刚刚';
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}分钟前`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}小时前`;
  if (diff < 7 * 86_400_000) return `${Math.floor(diff / 86_400_000)}天前`;
  return new Date(value).toLocaleDateString();
}

async function refreshUnreadCount() {
  try {
    const res = await getUnreadCountApi();
    unreadCount.value = res.count ?? 0;
  } catch {
    // 未读数获取失败不影响主流程
  }
}

async function loadNotifications() {
  try {
    const res = await getNotificationListApi({
      currentPage: 1,
      pageSize: 20,
    });
    notifications.value = (res.items ?? []).map((item) =>
      toNotificationItem(item),
    );
  } catch {
    notifications.value = [];
  }
  await refreshUnreadCount();
}

// SSE 实时推送：收到新通知直接插到列表头部并刷新未读数
let unsubscribeStream: () => void = () => {};
onMounted(() => {
  void loadNotifications();
  // 收到任何事件都重新拉取：新建与撤回都能覆盖，且未读数始终与后端一致
  unsubscribeStream = subscribeNotificationStream({
    onNotification: () => void loadNotifications(),
  });
});
onUnmounted(() => unsubscribeStream());

const router = useRouter();
const userStore = useUserStore();
const authStore = useAuthStore();
const accessStore = useAccessStore();
const { destroyWatermark, updateWatermark } = useWatermark();
// 用后端未读数判断，避免列表只取前 20 条时漏判
const showDot = computed(() => unreadCount.value > 0);

const profileOpen = ref(false);

const menus = computed(() => [
  {
    handler: () => {
      profileOpen.value = true;
    },
    icon: 'lucide:user',
    text: $t('page.auth.profile'),
  },
  {
    handler: () => {
      openWindow('http://weaveradmin.antsurge.com', {
        target: '_blank',
      });
    },
    icon: BookOpenText,
    text: $t('ui.widgets.document'),
  },
  {
    handler: () => {
      openWindow('https://github.com/antsurge/weaver-admin', {
        target: '_blank',
      });
    },
    icon: SvgGithubIcon,
    text: 'GitHub',
  },
  {
    handler: () => {
      openWindow('https://github.com/antsurge/weaver-admin/issues', {
        target: '_blank',
      });
    },
    icon: CircleHelp,
    text: $t('ui.widgets.qa'),
  },
]);

const avatar = computed(() => {
  return (
    getFileAccessUrl(userStore.userInfo?.avatar) ||
    preferences.app.defaultAvatar
  );
});

async function handleLogout() {
  await authStore.logout(false);
}

function handleNoticeClear() {
  const ids = notifications.value.map((item) => String(item.id));
  if (ids.length === 0) return;
  deleteNotificationApi(ids)
    .then(() => {
      notifications.value = [];
      return refreshUnreadCount();
    })
    .catch(() => {});
}

function markRead(id: number | string) {
  readNotificationApi(String(id))
    .then(() => {
      const item = notifications.value.find((item) => item.id === id);
      if (item) {
        item.isRead = true;
      }
      return refreshUnreadCount();
    })
    .catch(() => {});
}

function remove(id: number | string) {
  deleteNotificationApi([String(id)])
    .then(() => {
      notifications.value = notifications.value.filter(
        (item) => item.id !== id,
      );
      return refreshUnreadCount();
    })
    .catch(() => {});
}

function handleMakeAll() {
  readAllNotificationApi()
    .then(() => {
      notifications.value.forEach((item) => (item.isRead = true));
      unreadCount.value = 0;
    })
    .catch(() => {});
}

function handleViewAll() {
  router.push('/system/notification');
}
watch(
  () => ({
    enable: preferences.app.watermark,
    content: preferences.app.watermarkContent,
  }),
  async ({ enable, content }) => {
    if (enable) {
      await updateWatermark({
        content:
          content ||
          `${userStore.userInfo?.username} - ${userStore.userInfo?.realName}`,
      });
    } else {
      destroyWatermark();
    }
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <BasicLayout @clear-preferences-and-logout="handleLogout">
    <template #user-dropdown>
      <UserDropdown
        :avatar
        :menus
        :text="userStore.userInfo?.realName"
        description="ann.vben@gmail.com"
        tag-text="Pro"
        @logout="handleLogout"
      />
    </template>
    <template #notification>
      <Notification
        :dot="showDot"
        :notifications="notifications"
        @clear="handleNoticeClear"
        @read="(item) => item.id && markRead(item.id)"
        @remove="(item) => item.id && remove(item.id)"
        @make-all="handleMakeAll"
        @view-all="handleViewAll"
      />
    </template>
    <template #extra>
      <AuthenticationLoginExpiredModal
        v-model:open="accessStore.loginExpired"
        :avatar
      >
        <LoginForm />
      </AuthenticationLoginExpiredModal>
    </template>
    <template #lock-screen>
      <LockScreen :avatar @to-login="handleLogout" />
    </template>
  </BasicLayout>
  <ProfileModal v-model:open="profileOpen" />
</template>
