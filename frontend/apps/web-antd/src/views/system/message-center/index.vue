<script lang="ts" setup>
import type { NotificationApi } from '#/api/system/notification';

import { onMounted, onUnmounted, reactive, ref, watch } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import {
  Button,
  Empty,
  message,
  Modal,
  Pagination,
  Spin,
  Tag,
  Tooltip,
} from 'ant-design-vue';
import dayjs from 'dayjs';

import {
  deleteNotificationApi,
  getNotificationListApi,
  getUnreadCountApi,
  readAllNotificationApi,
  readNotificationApi,
  subscribeNotificationStream,
} from '#/api/system/notification';
import { $t } from '#/locales';
import { DEFAULT_PAGE_SIZE } from '#/types/pagination';

import {
  getMessageLevelColor,
  getMessageLevelLabel,
  getMessageTypeColor,
  getMessageTypeLabel,
  MESSAGE_TYPE_TABS,
} from './data';
import MessageDetailDrawer from './modules/detail/index.vue';

const activeType = ref('');

const list = ref<NotificationApi.NotificationRecord[]>([]);
const total = ref(0);
const unreadCount = ref(0);
const loading = ref(false);
const refreshing = ref(false);
const pageNo = ref(0);
const pageSize = DEFAULT_PAGE_SIZE;

const pagination = reactive({
  current: 1,
  pageSize: DEFAULT_PAGE_SIZE,
  total: 0,
});

const [DetailDrawer, drawerApi] = useVbenDrawer({
  connectedComponent: MessageDetailDrawer,
  destroyOnClose: true,
});

async function loadList(page = 1) {
  if (loading.value) return;

  loading.value = true;
  try {
    const params: NotificationApi.ListParams = {
      currentPage: page,
      pageSize,
    };
    if (activeType.value) {
      params.type = activeType.value;
    }

    const res = await getNotificationListApi(params);
    const items = res.items ?? [];

    list.value = items;
    total.value = res.total ?? 0;
    pageNo.value = page;
    pagination.current = page;
    pagination.total = total.value;
  } finally {
    loading.value = false;
    refreshing.value = false;
  }
}

async function refreshUnread() {
  try {
    const res = await getUnreadCountApi();
    unreadCount.value = res.count ?? 0;
  } catch {
    // 忽略未读数获取失败
  }
}

let unsubscribeStream: () => void = () => {};
onMounted(() => {
  void refreshUnread();
  void loadList(1);
  // SSE 实时推送：新消息到达自动刷新列表与未读数
  unsubscribeStream = subscribeNotificationStream({
    onNotification: () => {
      refreshing.value = true;
      void loadList(pagination.current);
      void refreshUnread();
    },
  });
});

onUnmounted(() => unsubscribeStream());

watch(activeType, () => {
  void loadList(1);
});

function showDetail(record: NotificationApi.NotificationRecord) {
  drawerApi.setData(record).open();

  // 打开详情时如未读则标记已读
  if (!record.isRead) {
    markRead(record.id);
  }
}

async function markRead(id: string) {
  try {
    await readNotificationApi(id);
    const item = list.value.find((v) => v.id === id);
    if (item) {
      item.isRead = true;
    }
    await refreshUnread();
  } catch {
    // 忽略标记已读失败
  }
}

async function handleReadAll() {
  if (unreadCount.value === 0) {
    message.info($t('system.message_center.message.noUnread'));
    return;
  }

  Modal.confirm({
    title: $t('system.message_center.confirm.readAllTitle'),
    content: $t('system.message_center.confirm.readAllContent', [
      String(unreadCount.value),
    ]),
    okText: $t('common.confirm'),
    cancelText: $t('common.cancel'),
    async onOk() {
      await readAllNotificationApi();
      list.value.forEach((item) => (item.isRead = true));
      await refreshUnread();
      message.success($t('system.message_center.message.readAllSuccess'));
    },
  });
}

async function handleDelete(ids: string[]) {
  if (ids.length === 0) {
    message.warning($t('system.message_center.message.selectDelete'));
    return;
  }
  Modal.confirm({
    title: $t('system.message_center.confirm.deleteTitle'),
    content: $t('system.message_center.confirm.deleteContent', [
      String(ids.length),
    ]),
    okText: $t('common.confirm'),
    cancelText: $t('common.cancel'),
    async onOk() {
      await deleteNotificationApi(ids);
      message.success($t('system.message_center.message.deleteSuccess'));
      selectedIds.value = [];
      await refreshUnread();
      // 当前页删空且非第一页时回退一页
      const targetPage =
        pagination.current > 1 && list.value.length === 0
          ? pagination.current - 1
          : pagination.current;
      void loadList(targetPage);
    },
  });
}

const selectedIds = ref<string[]>([]);

function onSelectionChange(id: string) {
  const idx = selectedIds.value.indexOf(id);
  if (idx === -1) {
    selectedIds.value.push(id);
  } else {
    selectedIds.value.splice(idx, 1);
  }
}

/** 消息卡片底色：未读带主题色边框与浅色背景 */
function rowClass(item: NotificationApi.NotificationRecord): string {
  return item.isRead
    ? 'border-border/70 bg-card'
    : 'border-primary/20 bg-primary/[0.04]';
}

/** 时间展示：今天显示时分，其余显示日期 */
function formatTime(value?: string): string {
  if (!value) return '';
  const t = dayjs(value);
  return t.isSame(dayjs(), 'day') ? t.format('HH:mm') : t.format('YYYY-MM-DD');
}
</script>

<template>
  <div
    class="flex h-full min-h-[540px] overflow-hidden rounded-md border border-border bg-card"
  >
    <!-- 左侧类型导航 -->
    <aside
      class="flex w-[168px] shrink-0 flex-col border-r border-border bg-muted/30 py-2.5"
    >
      <button
        v-for="tab in MESSAGE_TYPE_TABS"
        :key="tab.key"
        class="relative flex h-9 w-full cursor-pointer items-center justify-between px-4 text-sm transition-colors"
        :class="
          activeType === tab.key
            ? 'bg-primary/10 font-medium text-primary'
            : 'text-muted-foreground hover:bg-accent hover:text-foreground'
        "
        @click="activeType = tab.key"
      >
        <span
          v-if="activeType === tab.key"
          class="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-primary"
        ></span>
        <span class="truncate">{{ tab.label }}</span>
        <span
          v-if="tab.key === '' && unreadCount > 0"
          class="ml-2 shrink-0 rounded-full bg-red-500 px-1.5 text-xs !leading-[18px] text-white"
        >
          {{ unreadCount > 99 ? '99+' : unreadCount }}
        </span>
      </button>
    </aside>

    <!-- 右侧消息列表 -->
    <section class="flex min-w-0 flex-1 flex-col">
      <!-- 工具栏 -->
      <div
        class="flex shrink-0 items-center justify-between border-b border-border px-4 py-2.5"
      >
        <div class="flex items-center gap-2 text-sm">
          <IconifyIcon icon="lucide:bell" class="size-4" />
          <span class="font-medium">
            {{ $t('system.message_center.title') }}
          </span>
        </div>
        <div class="flex items-center gap-1">
          <span
            v-if="unreadCount > 0"
            class="mr-2 text-xs text-muted-foreground"
          >
            {{ $t('system.message_center.unreadCount', [unreadCount]) }}
          </span>
          <Tooltip :title="$t('system.message_center.operation.readAll')">
            <Button type="text" size="small" @click="handleReadAll">
              <IconifyIcon
                class="size-4"
                icon="ant-design:check-circle-outlined"
              />
            </Button>
          </Tooltip>
          <Tooltip :title="$t('system.message_center.operation.delete')">
            <Button type="text" size="small" @click="handleDelete(selectedIds)">
              <IconifyIcon class="size-4" icon="ant-design:delete-outlined" />
            </Button>
          </Tooltip>
        </div>
      </div>

      <!-- 列表 -->
      <div class="relative min-h-0 flex-1 overflow-y-auto">
        <Spin :spinning="loading">
          <Empty
            v-if="!loading && list.length === 0"
            class="py-24"
            :description="$t('system.message_center.empty')"
          />
          <ul
            v-else
            class="flex flex-col gap-1.5 p-2 transition-opacity"
            :class="refreshing ? 'opacity-60' : ''"
          >
            <li
              v-for="item in list"
              :key="item.id"
              class="group relative cursor-pointer"
              @click="showDetail(item)"
            >
              <!-- 消息卡片 -->
              <div
                class="relative flex items-start gap-3 rounded-lg border px-3.5 py-3.5 transition-all duration-200"
                :class="[
                  rowClass(item),
                  selectedIds.includes(item.id)
                    ? 'border-primary/40 bg-accent/70 shadow-sm'
                    : 'hover:border-primary/30 hover:bg-accent/50 hover:shadow-md',
                ]"
              >
                <!-- 多选框 -->
                <div
                  class="mt-0.5 shrink-0"
                  @click.stop="onSelectionChange(item.id)"
                >
                  <div
                    class="flex size-[16px] cursor-pointer items-center justify-center rounded-[4px] border transition-colors"
                    :class="
                      selectedIds.includes(item.id)
                        ? 'border-primary bg-primary'
                        : 'border-border bg-card group-hover:border-muted-foreground/40'
                    "
                  >
                    <IconifyIcon
                      v-if="selectedIds.includes(item.id)"
                      icon="ant-design:check-outlined"
                      class="size-3 text-white"
                    />
                  </div>
                </div>

                <!-- 消息主体 -->
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span
                      v-if="!item.isRead"
                      class="size-2 shrink-0 rounded-full bg-red-500"
                    ></span>
                    <span
                      class="truncate text-sm"
                      :class="
                        item.isRead
                          ? 'font-normal text-muted-foreground'
                          : 'font-semibold text-foreground'
                      "
                    >
                      {{ item.notification?.title }}
                    </span>
                    <span
                      class="ml-auto shrink-0 text-xs text-muted-foreground"
                    >
                      {{ formatTime(item.createdAt) }}
                    </span>
                  </div>
                  <!-- 内容区：浅色内容块，与主题明显区分 -->
                  <div
                    class="mt-1.5 line-clamp-2 rounded-md bg-muted/50 px-3 py-2 text-xs leading-5 text-muted-foreground transition-colors group-hover:bg-muted/80"
                  >
                    {{ item.notification?.content }}
                  </div>
                  <div class="mt-2 flex items-center gap-2">
                    <Tag
                      :color="
                        getMessageTypeColor(item.notification?.type ?? 'system')
                      "
                      class="!my-0 !text-xs"
                    >
                      {{
                        getMessageTypeLabel(item.notification?.type ?? 'system')
                      }}
                    </Tag>
                    <Tag
                      :color="
                        getMessageLevelColor(item.notification?.level ?? 'info')
                      "
                      class="!my-0 !text-xs"
                    >
                      {{
                        getMessageLevelLabel(item.notification?.level ?? 'info')
                      }}
                    </Tag>
                  </div>
                </div>
              </div>
            </li>
          </ul>
        </Spin>
      </div>

      <!-- 分页 -->
      <div
        v-if="total > 0"
        class="flex shrink-0 items-center justify-between border-t border-border px-4 py-2"
      >
        <span class="text-xs text-muted-foreground">
          {{ $t('system.message_center.pagination.total', [total]) }}
        </span>
        <Pagination
          v-model:current="pagination.current"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          :show-size-changer="false"
          size="small"
          @change="loadList"
        />
      </div>
    </section>

    <DetailDrawer @read="void loadList(pagination.current)" />
  </div>
</template>
