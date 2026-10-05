import type { NotificationApi } from '#/api/system/notification';

import { $t } from '#/locales';

/** 消息类型 Tab 定义（空值表示全部） */
export const MESSAGE_TYPE_TABS: Array<{
  color?: string;
  key: string;
  label: string;
}> = [
  { key: '', label: $t('system.message_center.tabs.all') },
  { key: 'system', label: $t('system.message_center.tabs.system') },
  { key: 'announce', label: $t('system.message_center.tabs.announce') },
  { key: 'audit', label: $t('system.message_center.tabs.audit') },
  { key: 'todo', label: $t('system.message_center.tabs.todo') },
  { key: 'alert', label: $t('system.message_center.tabs.alert') },
];

/** 消息类型选项（标签展示） */
export const MESSAGE_TYPE_OPTIONS = [
  {
    color: 'default',
    label: $t('system.message_center.types.system'),
    value: 'system',
  },
  {
    color: 'blue',
    label: $t('system.message_center.types.announce'),
    value: 'announce',
  },
  {
    color: 'purple',
    label: $t('system.message_center.types.audit'),
    value: 'audit',
  },
  {
    color: 'cyan',
    label: $t('system.message_center.types.todo'),
    value: 'todo',
  },
  {
    color: 'orange',
    label: $t('system.message_center.types.alert'),
    value: 'alert',
  },
];

/** 重要级别选项（标签展示） */
export const MESSAGE_LEVEL_OPTIONS = [
  { color: 'default', label: $t('common.info'), value: 'info' },
  { color: 'success', label: $t('common.success'), value: 'success' },
  { color: 'warning', label: $t('common.warning'), value: 'warn' },
  { color: 'error', label: $t('common.error'), value: 'error' },
];

export function getMessageTypeLabel(type: string): string {
  return (
    MESSAGE_TYPE_OPTIONS.find((item) => item.value === type)?.label ?? type
  );
}

export function getMessageLevelLabel(level: string): string {
  return (
    MESSAGE_LEVEL_OPTIONS.find((item) => item.value === level)?.label ?? level
  );
}

export function getMessageTypeColor(type: string): string {
  return (
    MESSAGE_TYPE_OPTIONS.find((item) => item.value === type)?.color ?? 'default'
  );
}

export function getMessageLevelColor(level: string): string {
  return (
    MESSAGE_LEVEL_OPTIONS.find((item) => item.value === level)?.color ??
    'default'
  );
}

/** 时间显示：支持相对时间 */
export function formatMessageTime(value?: string): string {
  if (!value) return '';
  return value;
}

export type MessageRecord = NotificationApi.NotificationRecord;
