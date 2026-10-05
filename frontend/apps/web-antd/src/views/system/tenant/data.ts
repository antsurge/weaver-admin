import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemTenantApi } from '#/api/system/tenant';

import { useAccess } from '@vben/access';

import { SystemAuthCode } from '#/access';
import { $t } from '#/locales';

export function useTenantColumns(
  onActionClick: OnActionClickFn<SystemTenantApi.Tenant>,
  onStatusChange?: (
    newStatus: SystemTenantApi.Tenant['status'],
    row: SystemTenantApi.Tenant,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemTenantApi.Tenant>['columns'] {
  const { hasAccessByCodes } = useAccess();
  return [
    {
      type: 'checkbox',
      width: 60,
      align: 'center',
    },
    {
      field: 'name',
      align: 'center',
      title: $t('system.tenant.fields.name'),
      minWidth: 140,
    },
    {
      field: 'code',
      align: 'center',
      title: $t('system.tenant.fields.code'),
      minWidth: 140,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.tenant.fields.status'),
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无启用/禁用按钮权限时禁用开关
          disabled: () =>
            !hasAccessByCodes([SystemAuthCode.Tenant.SwitchStatus]),
        },
      },
    },
    {
      field: 'maxUsers',
      align: 'center',
      title: $t('system.tenant.fields.maxUsers'),
      width: 110,
    },
    {
      field: 'maxRoles',
      align: 'center',
      title: $t('system.tenant.fields.maxRoles'),
      width: 110,
    },
    {
      field: 'contactName',
      align: 'center',
      title: $t('system.tenant.fields.contactName'),
      width: 110,
    },
    {
      field: 'expireAt',
      align: 'center',
      title: $t('system.tenant.fields.expireAt'),
      width: 180,
      formatter: 'formatDateTime',
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('system.tenant.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('system.tenant.fields.operation'),
      width: 160,
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
            // 无编辑权限时不展示
            show: hasAccessByCodes([SystemAuthCode.Tenant.Edit]),
          },
          {
            code: 'delete',
            // 无删除权限时不展示
            show: hasAccessByCodes([SystemAuthCode.Tenant.Delete]),
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
          placeholder: $t('system.tenant.searchPlaceholder.name'),
        },
        fieldName: 'name',
        label: $t('system.tenant.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('system.tenant.searchPlaceholder.code'),
        },
        fieldName: 'code',
        label: $t('system.tenant.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('system.tenant.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('system.tenant.fields.status'),
          ]),
          options: [
            { label: $t('common.enabled'), value: 'enabled' },
            { label: $t('common.disabled'), value: 'disabled' },
          ],
          allowClear: true,
        },
      },
    ],
    // 控制表单是否显示折叠按钮
    showCollapseButton: true,
    submitButtonOptions: {
      content: $t('common.actions.search'),
    },
    // 是否在字段值改变时提交表单
    submitOnChange: false,
    // 按下回车时是否提交表单
    submitOnEnter: false,
  };
}
