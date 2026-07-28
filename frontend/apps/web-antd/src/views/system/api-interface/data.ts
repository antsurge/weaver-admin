import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { VbenFormProps } from '#/adapter/form';
import type { SystemApiInterfaceApi } from '#/api/system/api-interface';
import { $t } from '#/locales';

export function useColumns(
  onActionClick: OnActionClickFn<SystemApiInterfaceApi.ApiInterface>,
): VxeTableGridOptions<SystemApiInterfaceApi.ApiInterface>['columns'] {
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
      field: 'service',
      align: 'center',
      title: $t('system.api_interface.fields.service'),
      width: 240,
    },
    {
      field: 'tag',
      align: 'center',
      title: $t('system.api_interface.fields.tag'),
      width: 120,
    },
    {
      field: 'method',
      align: 'center',
      title: $t('system.api_interface.fields.method'),
      width: 100,
    },
    {
      field: 'path',
      align: 'left',
      title: $t('system.api_interface.fields.path'),
      minWidth: 300,
    },
    {
      field: 'summary',
      align: 'left',
      title: $t('system.api_interface.fields.summary'),
      minWidth: 200,
      showOverflow: 'tooltip',
    },
    {
      field: 'createdAt',
      align: 'center',
      title: $t('common.fields.createdAt'),
      width: 180,
      formatter: ({ cellValue }) =>
        cellValue ? new Date(cellValue).toLocaleString() : '-',
    },
    {
      field: 'operation',
      align: 'center',
      fixed: 'right',
      showOverflow: false,
      title: $t('common.fields.operation'),
      width: 120,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'code',
          onClick: onActionClick,
        },
        options: ['delete'],
      },
    },
  ];
}

export function useGridFormOptions(): VbenFormProps {
  return {
    collapsed: false,
    schema: [
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.placeholder', [
            $t('system.api_interface.fields.service'),
          ]),
        },
        fieldName: 'service',
        label: $t('system.api_interface.fields.service'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.placeholder', [
            $t('system.api_interface.fields.tag'),
          ]),
        },
        fieldName: 'tag',
        label: $t('system.api_interface.fields.tag'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.placeholder', [
            $t('system.api_interface.fields.method'),
          ]),
        },
        fieldName: 'method',
        label: $t('system.api_interface.fields.method'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.placeholder', [
            $t('system.api_interface.fields.path'),
          ]),
        },
        fieldName: 'path',
        label: $t('system.api_interface.fields.path'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.placeholder', [
            $t('system.api_interface.fields.summary'),
          ]),
        },
        fieldName: 'summary',
        label: $t('system.api_interface.fields.summary'),
      },
    ],
    showCollapseButton: true,
    submitButtonOptions: {
      content: '查询',
    },
    submitOnChange: false,
    submitOnEnter: false,
  };
}