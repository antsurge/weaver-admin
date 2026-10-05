import type { Ref } from 'vue';

import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemConfigApi } from '#/api/system/config';

import { useAccess } from '@vben/access';

import { SystemAuthCode } from '#/access';
import { $t } from '#/locales';

const { hasAccessByCodes } = useAccess();

export function useConfigColumns(
  onActionClick: OnActionClickFn<SystemConfigApi.Config>,
  onStatusChange?: (
    newStatus: SystemConfigApi.Config['status'],
    row: SystemConfigApi.Config,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemConfigApi.Config>['columns'] {
  return [
    {
      type: 'checkbox',
      width: 60,
      align: 'center',
    },
    {
      field: 'name',
      align: 'center',
      title: $t('system.config.fields.name'),
      minWidth: 160,
    },
    {
      field: 'key',
      align: 'center',
      title: $t('system.config.fields.key'),
      minWidth: 180,
    },
    {
      field: 'group',
      align: 'center',
      title: $t('system.config.fields.group'),
      width: 130,
      showOverflow: true,
      slots: { default: 'config_group' },
    },
    {
      field: 'value',
      align: 'center',
      title: $t('system.config.fields.value'),
      minWidth: 200,
      showOverflow: true,
    },
    {
      field: 'configType',
      align: 'center',
      title: $t('system.config.fields.configType'),
      width: 110,
      slots: { default: 'config_type' },
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.config.fields.status'),
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无启用/禁用权限时禁用开关
          disabled: () =>
            !hasAccessByCodes([SystemAuthCode.Config.SwitchStatus]),
        },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('system.config.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('system.config.fields.operation'),
      width: 200,
      fixed: 'right',
      showOverflow: false,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'name',
          onClick: onActionClick,
        },
        options: [
          {
            code: 'edit',
            // 无编辑按钮权限时不展示
            show: hasAccessByCodes([SystemAuthCode.Config.Edit]),
          },
          {
            code: 'delete',
            // 无删除按钮权限时不展示
            show: hasAccessByCodes([SystemAuthCode.Config.Delete]),
          },
        ],
      },
    },
  ];
}

export function useFormOptions(
  groupOptions: Ref<{ label: string; value: string }[]>,
): VbenFormProps {
  return {
    collapsed: false,
    schema: [
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('system.config.search_placeholder.name'),
        },
        fieldName: 'name',
        label: $t('system.config.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('system.config.search_placeholder.key'),
        },
        fieldName: 'key',
        label: $t('system.config.fields.key'),
      },
      {
        component: 'Select',
        fieldName: 'group',
        label: $t('system.config.fields.group'),
        componentProps: () => ({
          placeholder: $t('system.config.search_placeholder.group'),
          allowClear: true,
          showSearch: true,
          options: groupOptions.value,
        }),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('system.config.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('system.config.fields.status'),
          ]),
          options: [
            { label: $t('common.enabled'), value: 'enabled' },
            { label: $t('common.disabled'), value: 'disabled' },
          ],
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
