import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemDataPermissionApi } from '#/api/permission/data-permission';

import { useAccess } from '@vben/access';

import { PermissionAuthCode } from '#/access';
import { $t } from '#/locales';

export const SCOPE_TYPE_OPTIONS = [
  { label: $t('permission.data_permission.scope.all'), value: '1' },
  { label: $t('permission.data_permission.scope.custom'), value: '2' },
  { label: $t('permission.data_permission.scope.dept'), value: '3' },
  { label: $t('permission.data_permission.scope.deptAndBelow'), value: '4' },
  { label: $t('permission.data_permission.scope.self'), value: '5' },
];

export function formatScopeType(scopeType: string): string {
  return (
    SCOPE_TYPE_OPTIONS.find((item) => item.value === scopeType)?.label ??
    scopeType
  );
}

export function useDataPermissionColumns(
  onActionClick: OnActionClickFn<SystemDataPermissionApi.DataPermission>,
  onStatusChange?: (
    newStatus: SystemDataPermissionApi.DataPermission['status'],
    row: SystemDataPermissionApi.DataPermission,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemDataPermissionApi.DataPermission>['columns'] {
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
      title: $t('permission.data_permission.fields.name'),
      minWidth: 140,
    },
    {
      field: 'code',
      align: 'center',
      title: $t('permission.data_permission.fields.code'),
      width: 180,
    },
    {
      field: 'scopeType',
      align: 'center',
      title: $t('permission.data_permission.fields.scopeType'),
      width: 160,
      slots: { default: 'scope_type' },
    },
    {
      field: 'status',
      align: 'center',
      title: $t('permission.data_permission.fields.status'),
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无启用/禁用按钮权限时禁用开关
          disabled: () =>
            !hasAccessByCodes([PermissionAuthCode.DataPermission.SwitchStatus]),
        },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('permission.data_permission.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('permission.data_permission.fields.operation'),
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
            // 无编辑按钮权限时不展示
            show: hasAccessByCodes([PermissionAuthCode.DataPermission.Edit]),
          },
          {
            code: 'delete',
            // 无删除按钮权限时不展示
            show: hasAccessByCodes([PermissionAuthCode.DataPermission.Delete]),
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
          placeholder: $t('ui.formRules.required', [
            $t('permission.data_permission.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('permission.data_permission.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('permission.data_permission.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('permission.data_permission.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'scopeType',
        label: $t('permission.data_permission.fields.scopeType'),
        componentProps: {
          class: 'w-full',
          placeholder: $t('ui.formRules.selectRequired', [
            $t('permission.data_permission.fields.scopeType'),
          ]),
          options: SCOPE_TYPE_OPTIONS,
          allowClear: true,
        },
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('permission.data_permission.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('permission.data_permission.fields.status'),
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
