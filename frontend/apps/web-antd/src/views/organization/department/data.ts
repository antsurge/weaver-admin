import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { OrganizationDepartmentApi } from '#/api/organization/department';
import type { VbenFormProps } from '#/adapter/form';
import { $t } from '#/locales';

export function useColumns(
  onActionClick: OnActionClickFn<OrganizationDepartmentApi.Department>,
  onStatusChange?: (
    newStatus: OrganizationDepartmentApi.Department['status'],
    row: OrganizationDepartmentApi.Department,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<OrganizationDepartmentApi.Department>['columns'] {
  return [
    {
      type: 'checkbox',
      align: 'center',
      width: 80,
      fixed: 'left',
    },
    {
      field: 'name',
      align: 'center',
      title: $t('organization.department.fields.name'),
      treeNode: true,
    },
    {
      field: 'code',
      align: 'center',
      title: $t('organization.department.fields.code'),
      width: 200,
    },
    {
      field: 'weight',
      align: 'center',
      title: $t('organization.department.fields.weight'),
      width: 120,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('organization.department.fields.status'),
      width: 140,
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('organization.department.fields.updatedAt'),
      width: 200,
    },
    {
      field: 'operation',
      align: 'center',
      fixed: 'right',
      showOverflow: false,
      title: $t('organization.department.fields.operation'),
      width: 260,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'name',
          onClick: onActionClick,
        },
        options: [
          {
            code: 'append',
            text: $t('organization.department.operation.appendChildren'),
          },
          'edit',
          'delete',
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
            $t('organization.department.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('organization.department.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('organization.department.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('organization.department.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('organization.department.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('organization.department.fields.status'),
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
