import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemOrganizationApi } from '#/api/system/organization';
import type { VbenFormProps } from '#/adapter/form';
import { $t } from '#/locales';

export function useColumns(
  onActionClick: OnActionClickFn<SystemOrganizationApi.Organization>,
  onStatusChange?: (
    newStatus: SystemOrganizationApi.Organization['status'],
    row: SystemOrganizationApi.Organization,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemOrganizationApi.Organization>['columns'] {
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
      title: $t('system.organization.fields.name'),
      treeNode: true,
    },
    {
      field: 'code',
      align: 'center',
      title: $t('system.organization.fields.code'),
      width: 200,
    },
    {
      field: 'weight',
      align: 'center',
      title: $t('system.organization.fields.weight'),
      width: 120,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.organization.fields.status'),
      width: 140,
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('system.organization.fields.updatedAt'),
      width: 200,
    },
    {
      field: 'operation',
      align: 'center',
      fixed: 'right',
      showOverflow: false,
      title: $t('system.organization.fields.operation'),
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
            text: $t('system.organization.operation.appendChildren'),
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
            $t('system.organization.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('system.organization.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('system.organization.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('system.organization.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('system.organization.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('system.organization.fields.status'),
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
