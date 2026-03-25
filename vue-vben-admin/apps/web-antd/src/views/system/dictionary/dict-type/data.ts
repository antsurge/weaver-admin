import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { DictionaryDictTypeApi } from '#/api/system/dictionary/dict-type';
import type { VbenFormProps } from '#/adapter/form';

import { $t } from '#/locales';

export function useDictTypeColumns(
  onActionClick: OnActionClickFn<DictionaryDictTypeApi.DictType>,
  onStatusChange?: (
    newStatus: DictionaryDictTypeApi.DictType['status'],
    row: DictionaryDictTypeApi.DictType,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<DictionaryDictTypeApi.DictType>['columns'] {
  return [
    {
      type: 'expand',
      width: 60,
      align: 'center',
      slots: { content: "expand_dictdata" }
    },
    {
      field: 'name',
      align: 'center',
      title: $t('system.dict_type.fields.name'),
    },
    {
      field: 'code',
      align: 'center',
      title: $t('system.dict_type.fields.code'),
      width: 200,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.dict_type.fields.status'),
      width: 140,
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('system.dict_type.fields.updatedAt'),
      width: 200,
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('system.dict_type.fields.operation'),
      width: 260,
      fixed: 'right',
      showOverflow: false,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'name',
          onClick: onActionClick,
        },
        options: [
          {
            code: 'appendDictData',
            text: $t('system.dict_type.operation.appendDictData'),
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
            $t('organization.position.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('organization.position.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('organization.position.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('organization.position.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('organization.position.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('organization.position.fields.status'),
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
