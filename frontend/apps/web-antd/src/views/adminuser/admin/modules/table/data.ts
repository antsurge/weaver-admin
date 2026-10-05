import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { AdminuserAdminApi } from '#/api/adminuser/admin';

import { useAccess } from '@vben/access';

import { AdminAuthCode } from '#/access';
import { $t } from '#/locales';

const { hasAccessByCodes } = useAccess();

export function useColumns(
  onActionClick: OnActionClickFn<AdminuserAdminApi.Admin>,
  onStatusChange?: (
    newStatus: AdminuserAdminApi.Admin['status'],
    row: AdminuserAdminApi.Admin,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<AdminuserAdminApi.Admin>['columns'] {
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
      field: 'avatar',
      align: 'center',
      title: $t('adminuser.admin.fields.avatar'),
      width: 80,
      slots: { default: 'avatar' },
    },
    {
      field: 'username',
      align: 'center',
      title: $t('adminuser.admin.fields.username'),
      width: 180,
    },
    {
      field: 'realName',
      align: 'center',
      title: $t('adminuser.admin.fields.realName'),
      width: 180,
    },
    {
      field: 'phone',
      align: 'center',
      title: $t('adminuser.admin.fields.phone'),
      width: 160,
    },
    {
      field: 'email',
      align: 'center',
      title: $t('adminuser.admin.fields.email'),
      width: 200,
    },
    {
      field: 'departmentName',
      align: 'center',
      title: $t('adminuser.admin.fields.departmentName'),
      width: 140,
      showOverflow: 'tooltip',
    },
    {
      field: 'roleNames',
      align: 'center',
      title: $t('adminuser.admin.fields.roleIds'),
      width: 220,
      showOverflow: 'tooltip',
      slots: { default: 'roleNames' },
    },
    {
      field: 'dataPermissionNames',
      align: 'center',
      title: $t('adminuser.admin.fields.dataPermissionIds'),
      width: 200,
      slots: { default: 'dataScope' },
    },
    {
      field: 'status',
      align: 'center',
      title: $t('adminuser.admin.fields.status'),
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无启用/禁用权限，或超级管理员账号时禁用开关
          disabled: (row: AdminuserAdminApi.Admin) =>
            !hasAccessByCodes([AdminAuthCode.Admin.SwitchStatus]) ||
            !!row.isSuperAdmin,
        },
      },
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
      width: 260,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'username',
          onClick: onActionClick,
        },
        options: [
          {
            code: 'edit',
            text: $t('ui.actionTitle.edit'),
            show: hasAccessByCodes([AdminAuthCode.Admin.Edit]),
          },
          {
            code: 'resetPassword',
            text: $t('adminuser.admin.actions.resetPassword'),
            show: (row: AdminuserAdminApi.Admin) =>
              hasAccessByCodes([AdminAuthCode.Admin.ResetPassword]) &&
              !row.isSuperAdmin,
          },
          {
            code: 'delete',
            text: $t('ui.actionTitle.delete'),
            // 超级管理员账号不展示删除按钮
            show: (row: AdminuserAdminApi.Admin) =>
              hasAccessByCodes([AdminAuthCode.Admin.Delete]) &&
              !row.isSuperAdmin,
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
            $t('adminuser.admin.fields.username'),
          ]),
        },
        fieldName: 'username',
        label: $t('adminuser.admin.fields.username'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('adminuser.admin.fields.realName'),
          ]),
        },
        fieldName: 'realName',
        label: $t('adminuser.admin.fields.realName'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('adminuser.admin.fields.phone'),
          ]),
        },
        fieldName: 'phone',
        label: $t('adminuser.admin.fields.phone'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('adminuser.admin.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('adminuser.admin.fields.status'),
          ]),
          options: [
            { label: '启用', value: 'enabled' },
            { label: '禁用', value: 'disabled' },
          ],
          allowClear: true,
        },
      },
    ],
    // 控制表单是否显示折叠按钮
    showCollapseButton: true,
    submitButtonOptions: {
      content: '查询',
    },
    // 是否在字段值改变时提交表单
    submitOnChange: false,
    // 按下回车时是否提交表单
    submitOnEnter: false,
  };
}
