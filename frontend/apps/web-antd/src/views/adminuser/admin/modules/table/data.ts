import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { VbenFormProps } from '#/adapter/form';
import type { OrganizationPositionApi } from '#/api/organization/position';

import { $t } from '#/locales';

import { useAccess } from '@vben/access';

const { hasAccessByCodes } = useAccess();

export function useColumns(
  onActionClick: OnActionClickFn<OrganizationPositionApi.Position>,
  onStatusChange?: (
    newStatus: OrganizationPositionApi.Position['status'],
    row: OrganizationPositionApi.Position,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<OrganizationPositionApi.Position>['columns'] {
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
      field: 'username',
      align: 'center',
      title: $t('adminuser.admin.fields.username'),
      width:180,
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
      field: 'status',
      align: 'center',
      title: $t('adminuser.admin.fields.status'),
      width: 120,
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
    },
    {
      field: 'createdAt',
      align: 'center',
      title: $t('common.fields.createdAt'),
      width: 180,
      formatter: ({ cellValue }) => cellValue ? new Date(cellValue).toLocaleString() : '-', // 格式化
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
          onClick: onActionClick,
        },
        // options: ['edit', 'delete'],
        options: [
          {
            code: 'edit',
            show: hasAccessByCodes(['Adminuser:Admin:Edit']),
          },
          {
            code: 'delete',
            show: hasAccessByCodes(['Adminuser:Admin:Delete']),
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
        fieldName: 'name',
        label: $t('adminuser.admin.fields.username'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('adminuser.admin.fields.realName'),
          ]),
        },
        fieldName: 'code',
        label: $t('adminuser.admin.fields.realName'),
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
  }
}
