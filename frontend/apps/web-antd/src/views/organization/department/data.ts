import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { OrganizationDepartmentApi } from '#/api/organization/department';

import { useAccess } from '@vben/access';

import { PermissionAuthCode } from '#/access';
import { $t } from '#/locales';

export function useColumns(
  onActionClick: OnActionClickFn<OrganizationDepartmentApi.Department>,
  onStatusChange?: (
    newStatus: OrganizationDepartmentApi.Department['status'],
    row: OrganizationDepartmentApi.Department,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<OrganizationDepartmentApi.Department>['columns'] {
  const { hasAccessByCodes } = useAccess();

  return [
    {
      type: 'checkbox',
      align: 'center',
      width: 60,
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
      field: 'type',
      align: 'center',
      title: $t('organization.department.fields.type'),
      width: 110,
      cellRender: {
        name: 'CellTag',
        options: [
          {
            value: 'company',
            color: 'purple',
            label: $t('organization.department.type.company'),
          },
          {
            value: 'subsidiary',
            color: 'success',
            label: $t('organization.department.type.subsidiary'),
          },
          {
            value: 'department',
            color: 'blue',
            label: $t('organization.department.type.department'),
          },
          {
            value: 'position',
            color: 'warning',
            label: $t('organization.department.type.position'),
          },
        ],
      },
    },
    {
      field: 'weight',
      align: 'center',
      title: $t('organization.department.fields.weight'),
      width: 80,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('organization.department.fields.status'),
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无编辑/切换状态权限时禁用开关
          disabled: () =>
            !hasAccessByCodes([PermissionAuthCode.Department.Edit]),
        },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('organization.department.fields.updatedAt'),
      width: 160,
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
            // 无新增下级权限时不展示
            show: hasAccessByCodes([PermissionAuthCode.Department.Append]),
          },
          {
            code: 'edit',
            // 无编辑权限时不展示
            show: hasAccessByCodes([PermissionAuthCode.Department.Edit]),
          },
          {
            code: 'delete',
            // 无删除权限时不展示
            show: hasAccessByCodes([PermissionAuthCode.Department.Delete]),
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
  };
}
