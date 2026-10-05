import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { FormSchemaApi } from '#/api/lowcode/form-schema';

import { $t } from '#/locales';

export function useFormSchemaColumns(
  onActionClick: OnActionClickFn<FormSchemaApi.FormSchema>,
): VxeTableGridOptions<FormSchemaApi.FormSchema>['columns'] {
  return [
    {
      type: 'checkbox',
      width: 60,
      align: 'center',
    },
    {
      field: 'name',
      align: 'center',
      title: $t('lowcode.formBuilder.fields.name'),
      minWidth: 160,
    },
    {
      field: 'code',
      align: 'center',
      title: $t('lowcode.formBuilder.fields.code'),
      minWidth: 180,
    },
    {
      field: 'description',
      align: 'center',
      title: $t('lowcode.formBuilder.fields.description'),
      minWidth: 220,
      showOverflow: true,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('lowcode.formBuilder.fields.status'),
      width: 100,
      slots: { default: 'status' },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('lowcode.formBuilder.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('lowcode.formBuilder.fields.operation'),
      width: 280,
      fixed: 'right',
      showOverflow: false,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'name',
          onClick: onActionClick,
        },
        options: [
          'edit',
          'design',
          {
            code: 'preview',
            text: $t('lowcode.formBuilder.operation.preview'),
          },
          {
            code: 'submissions',
            text: $t('lowcode.formBuilder.operation.submissions'),
          },
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
          placeholder: $t('lowcode.formBuilder.search_placeholder.name'),
        },
        fieldName: 'name',
        label: $t('lowcode.formBuilder.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('lowcode.formBuilder.search_placeholder.code'),
        },
        fieldName: 'code',
        label: $t('lowcode.formBuilder.fields.code'),
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
