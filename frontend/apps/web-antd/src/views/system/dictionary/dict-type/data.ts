import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { DictionaryDictTypeApi } from '#/api/system/dictionary/dict-type';

import { useAccess } from '@vben/access';

import { SystemAuthCode } from '#/access';
import { $t } from '#/locales';

export function useDictTypeColumns(
  onActionClick: OnActionClickFn<DictionaryDictTypeApi.DictType>,
  onStatusChange?: (
    newStatus: DictionaryDictTypeApi.DictType['status'],
    row: DictionaryDictTypeApi.DictType,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<DictionaryDictTypeApi.DictType>['columns'] {
  const { hasAccessByCodes } = useAccess();
  return [
    {
      type: 'expand',
      width: 60,
      align: 'center',
      slots: { content: 'expand_dictdata' },
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
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无编辑权限时禁用状态开关
          disabled: () => !hasAccessByCodes([SystemAuthCode.DictType.Edit]),
        },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('system.dict_type.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
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
            // 无新增字典数据权限时不展示
            show: hasAccessByCodes([SystemAuthCode.DictType.AddSubordinate]),
          },
          {
            code: 'edit',
            // 无编辑权限时不展示
            show: hasAccessByCodes([SystemAuthCode.DictType.Edit]),
          },
          {
            code: 'delete',
            // 无删除权限时不展示
            show: hasAccessByCodes([SystemAuthCode.DictType.Delete]),
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
            $t('system.dict_type.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('system.dict_type.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('system.dict_type.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('system.dict_type.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('system.dict_type.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('system.dict_type.fields.status'),
          ]),
          options: [
            { label: $t('common.enabled'), value: 'enabled' },
            { label: $t('common.disabled'), value: 'disabled' },
          ],
          allowClear: true,
        },
      },
    ],
    // 控制表单是否显示折叠按钮
    showCollapseButton: true,
    submitButtonOptions: {
      content: $t('common.actions.search'),
    },
    // 是否在字段值改变时提交表单
    submitOnChange: false,
    // 按下回车时是否提交表单
    submitOnEnter: false,
  };
}
