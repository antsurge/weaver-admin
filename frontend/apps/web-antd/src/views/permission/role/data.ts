import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { PermissionRoleApi } from '#/api/permission/role';

import { useAccess } from '@vben/access';

import { PermissionAuthCode } from '#/access';
import { $t } from '#/locales';

export function useColumns(
  onActionClick: OnActionClickFn<PermissionRoleApi.Role>,
  onStatusChange?: (
    newStatus: PermissionRoleApi.Role['status'],
    row: PermissionRoleApi.Role,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<PermissionRoleApi.Role>['columns'] {
  const { hasAccessByCodes } = useAccess();
  return [
    {
      type: 'checkbox',
      align: 'center',
      width: 60,
      fixed: 'left',
    },
    {
      type: 'seq',
      title: $t('common.fields.seq'),
      width: 60,
      align: 'center',
    },
    {
      field: 'name',
      align: 'center',
      title: $t('permission.role.fields.name'),
      width: 160,
    },
    {
      field: 'code',
      align: 'center',
      title: $t('permission.role.fields.code'),
      width: 220,
    },
    {
      field: 'isSystem',
      align: 'center',
      title: $t('permission.role.fields.isSystem'),
      width: 120,
      cellRender: {
        name: 'CellTag',
        options: [
          {
            value: true,
            color: 'warning',
            label: $t('permission.role.system'),
          },
          { value: false, color: 'blue', label: $t('permission.role.custom') },
        ],
      },
    },
    {
      field: 'weight',
      align: 'center',
      title: $t('permission.role.fields.weight'),
      width: 80,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('permission.role.fields.status'),
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无启用/禁用按钮权限时禁用开关
          disabled: () =>
            !hasAccessByCodes([PermissionAuthCode.Role.SwitchStatus]),
        },
      },
    },
    {
      field: 'createdAt',
      align: 'center',
      title: $t('common.fields.createdAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('common.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'operation',
      align: 'center',
      fixed: 'right',
      showOverflow: false,
      title: $t('common.fields.operation'),
      width: 200,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'name',
          // edit / delete 按钮对系统内置角色禁用
          onClick: onActionClick,
        },
        options: [
          {
            code: 'edit',
            // 无编辑按钮权限时不展示
            show: hasAccessByCodes([PermissionAuthCode.Role.Edit]),
            disabled: (row: PermissionRoleApi.Role) =>
              !!row.isSystem || !!row.isSuperAdmin,
          },
          {
            code: 'delete',
            // 超级管理员角色不展示删除按钮；系统内置角色禁用删除；无删除按钮权限不展示
            show: (row: PermissionRoleApi.Role) =>
              !row.isSuperAdmin &&
              hasAccessByCodes([PermissionAuthCode.Role.Delete]),
            disabled: (row: PermissionRoleApi.Role) =>
              !!row.isSystem || !!row.isSuperAdmin,
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
            $t('permission.role.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('permission.role.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('permission.role.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('permission.role.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('permission.role.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('permission.role.fields.status'),
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
