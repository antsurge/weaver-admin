import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { VbenFormProps } from '#/adapter/form';
import type { SystemPositionApi } from '#/api/system/position';

import { $t } from '#/locales';

export function useColumns(
  onActionClick: OnActionClickFn<SystemPositionApi.Position>,
  onStatusChange?: (
    newStatus: SystemPositionApi.Position['status'],
    row: SystemPositionApi.Position,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemPositionApi.Position>['columns'] {
  return [
    {
      type: 'checkbox',
      align: 'center',
      width: 80,
      fixed: 'left',
    },
    {
      type: 'seq',
      title: $t('common.fields.seq'),
      width: 80,
      align: 'center', 
    },
    {
      field: 'name',
      align: 'center',
      title: $t('system.position.fields.name'),
    },
    {
      field: 'code',
      align: 'center',
      title: $t('system.position.fields.code'),
      width: 260,
    },
    {
      field: 'weight',
      align: 'center',
      title: $t('system.position.fields.weight'),
      width: 160,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.position.fields.status'),
      width: 140,
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
        options: ['edit', 'delete'],
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
            $t('system.position.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('system.position.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('system.position.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('system.position.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('system.position.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('system.position.fields.status'),
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
