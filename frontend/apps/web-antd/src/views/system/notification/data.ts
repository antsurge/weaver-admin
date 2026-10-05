import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { NotificationApi } from '#/api/system/notification';

import { useAccess } from '@vben/access';

import { SystemAuthCode } from '#/access';
import { $t } from '#/locales';

/** 通知类型选项 */
export function getNotificationTypeOptions() {
  return [
    {
      color: 'default',
      label: $t('system.notification.type_options.system'),
      value: 'system',
    },
    {
      color: 'blue',
      label: $t('system.notification.type_options.announce'),
      value: 'announce',
    },
    {
      color: 'purple',
      label: $t('system.notification.type_options.audit'),
      value: 'audit',
    },
    {
      color: 'cyan',
      label: $t('system.notification.type_options.todo'),
      value: 'todo',
    },
    {
      color: 'orange',
      label: $t('system.notification.type_options.alert'),
      value: 'alert',
    },
  ];
}

/** 重要级别选项 */
export function getNotificationLevelOptions() {
  return [
    {
      color: 'default',
      label: $t('system.notification.level_options.info'),
      value: 'info',
    },
    {
      color: 'success',
      label: $t('system.notification.level_options.success'),
      value: 'success',
    },
    {
      color: 'warning',
      label: $t('system.notification.level_options.warn'),
      value: 'warn',
    },
    {
      color: 'error',
      label: $t('system.notification.level_options.error'),
      value: 'error',
    },
  ];
}

/** 投放范围选项 */
export function getNotificationTargetOptions() {
  return [
    {
      color: 'blue',
      label: $t('system.notification.target_options.all'),
      value: 'all',
    },
    {
      color: 'green',
      label: $t('system.notification.target_options.user'),
      value: 'user',
    },
    {
      color: 'purple',
      label: $t('system.notification.target_options.role'),
      value: 'role',
    },
  ];
}

/** 状态选项 */
export function getNotificationStatusOptions() {
  return [
    {
      color: 'default',
      label: $t('system.notification.status_options.draft'),
      value: 'draft',
    },
    {
      color: 'success',
      label: $t('system.notification.status_options.published'),
      value: 'published',
    },
    {
      color: 'error',
      label: $t('system.notification.status_options.revoked'),
      value: 'revoked',
    },
  ];
}

export function useNotificationColumns(
  onActionClick: OnActionClickFn<NotificationApi.Notification>,
): VxeTableGridOptions<NotificationApi.Notification>['columns'] {
  const { hasAccessByCodes } = useAccess();
  return [
    {
      field: 'title',
      align: 'center',
      title: $t('system.notification.fields.title'),
      minWidth: 180,
      showOverflow: 'tooltip',
    },
    {
      field: 'content',
      align: 'center',
      title: $t('system.notification.fields.content'),
      minWidth: 240,
      showOverflow: 'tooltip',
    },
    {
      field: 'type',
      align: 'center',
      title: $t('system.notification.fields.type'),
      width: 100,
      cellRender: { name: 'CellTag', options: getNotificationTypeOptions() },
    },
    {
      field: 'level',
      align: 'center',
      title: $t('system.notification.fields.level'),
      width: 100,
      cellRender: { name: 'CellTag', options: getNotificationLevelOptions() },
    },
    {
      field: 'targetType',
      align: 'center',
      title: $t('system.notification.fields.targetType'),
      width: 120,
      cellRender: { name: 'CellTag', options: getNotificationTargetOptions() },
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.notification.fields.status'),
      width: 100,
      cellRender: { name: 'CellTag', options: getNotificationStatusOptions() },
    },
    {
      field: 'createdAt',
      align: 'center',
      title: $t('system.notification.fields.createdAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('system.notification.fields.operation'),
      width: 120,
      fixed: 'right',
      showOverflow: false,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'title',
          onClick: onActionClick,
        },
        options: [
          {
            code: 'revoke',
            text: $t('system.notification.operation.revoke'),
            // 无撤回权限时不展示
            show: hasAccessByCodes([SystemAuthCode.Notification.Revoke]),
          },
        ],
      },
    },
  ];
}

export function useFormOptions(): VbenFormProps {
  return {
    collapsed: false,
    schema: [
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('system.notification.form_placeholder.title'),
        },
        fieldName: 'title',
        label: $t('system.notification.fields.title'),
      },
      {
        component: 'Select',
        fieldName: 'type',
        label: $t('system.notification.fields.type'),
        componentProps: {
          options: getNotificationTypeOptions(),
          allowClear: true,
        },
      },
      {
        component: 'Select',
        fieldName: 'level',
        label: $t('system.notification.fields.level'),
        componentProps: {
          options: getNotificationLevelOptions(),
          allowClear: true,
        },
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('system.notification.fields.status'),
        componentProps: {
          options: getNotificationStatusOptions(),
          allowClear: true,
        },
      },
    ],
    showCollapseButton: true,
    submitButtonOptions: {
      content: $t('common.actions.search'),
    },
    submitOnChange: false,
    submitOnEnter: false,
  };
}
