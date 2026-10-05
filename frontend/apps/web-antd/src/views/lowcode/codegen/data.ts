import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { CodegenApi } from '#/api/lowcode/codegen';

import { $t } from '#/locales';

export function useCodegenColumns(
  onActionClick: OnActionClickFn<CodegenApi.GenTable>,
): VxeTableGridOptions<CodegenApi.GenTable>['columns'] {
  return [
    {
      field: 'tableName',
      align: 'center',
      title: $t('lowcode.codegen.fields.tableName'),
      minWidth: 160,
    },
    {
      field: 'tableComment',
      align: 'center',
      title: $t('lowcode.codegen.fields.tableComment'),
      minWidth: 160,
    },
    {
      field: 'moduleName',
      align: 'center',
      title: $t('lowcode.codegen.fields.moduleName'),
      width: 120,
    },
    {
      field: 'bizName',
      align: 'center',
      title: $t('lowcode.codegen.fields.bizName'),
      width: 140,
    },
    {
      field: 'genType',
      align: 'center',
      title: $t('lowcode.codegen.fields.genType'),
      width: 110,
      slots: { default: 'gen_type' },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('lowcode.codegen.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('lowcode.codegen.fields.operation'),
      width: 220,
      fixed: 'right',
      showOverflow: false,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'tableName',
          onClick: onActionClick,
        },
        options: [
          { code: 'copy', text: $t('lowcode.codegen.operation.copy') },
          {
            code: 'delete',
            danger: true,
            text: $t('lowcode.codegen.operation.deleteRecord'),
          },
          {
            code: 'deleteModule',
            danger: true,
            text: $t('lowcode.codegen.operation.deleteModule'),
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
          placeholder: $t('lowcode.codegen.search_placeholder.tableName'),
        },
        fieldName: 'tableName',
        label: $t('lowcode.codegen.fields.tableName'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('lowcode.codegen.search_placeholder.bizName'),
        },
        fieldName: 'bizName',
        label: $t('lowcode.codegen.fields.bizName'),
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
