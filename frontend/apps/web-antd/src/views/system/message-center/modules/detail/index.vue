<script lang="ts" setup>
import type { NotificationApi } from '#/api/system/notification';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Descriptions, Tag } from 'ant-design-vue';

import { $t } from '#/locales';

import {
  getMessageLevelColor,
  getMessageLevelLabel,
  getMessageTypeColor,
  getMessageTypeLabel,
} from '../../data';

defineOptions({ name: 'MessageDetailDrawer' });

const emits = defineEmits<{ read: [void] }>();

const [Drawer, drawerApi] = useVbenDrawer({
  onOpenChange(isOpen) {
    if (isOpen) {
      record.value =
        drawerApi.getData<NotificationApi.NotificationRecord>() ?? undefined;
    }
  },
});

const record = ref<NotificationApi.NotificationRecord | undefined>();

const notification = computed(() => record.value?.notification);

const title = computed(() => notification.value?.title ?? '');
const content = computed(() => notification.value?.content ?? '');
const senderName = computed(() => {
  return (
    notification.value?.senderName ||
    $t('system.message_center.fields.systemSender')
  );
});

function handleRead() {
  if (record.value && !record.value.isRead) {
    emits('read');
  }
}
</script>

<template>
  <Drawer
    :title="title"
    width="520"
    class="message-detail-drawer"
    @close="handleRead"
  >
    <Descriptions v-if="notification" :column="1" bordered size="small">
      <Descriptions.Item :label="$t('system.message_center.fields.type')">
        <Tag :color="getMessageTypeColor(notification.type)">
          {{ getMessageTypeLabel(notification.type) }}
        </Tag>
      </Descriptions.Item>
      <Descriptions.Item :label="$t('system.message_center.fields.level')">
        <Tag :color="getMessageLevelColor(notification.level)">
          {{ getMessageLevelLabel(notification.level) }}
        </Tag>
      </Descriptions.Item>
      <Descriptions.Item :label="$t('system.message_center.fields.sender')">
        {{ senderName }}
      </Descriptions.Item>
      <Descriptions.Item :label="$t('system.message_center.fields.time')">
        {{ notification.createdAt || '-' }}
      </Descriptions.Item>
      <Descriptions.Item :label="$t('system.message_center.fields.content')">
        <div class="whitespace-pre-wrap">{{ content }}</div>
      </Descriptions.Item>
    </Descriptions>
  </Drawer>
</template>
