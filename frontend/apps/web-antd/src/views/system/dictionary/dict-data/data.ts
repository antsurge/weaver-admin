import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import { $t } from '#/locales';

import type { DictionaryDictDataApi } from '#/api/system/dictionary/dict-data';
export function useDictDataColumns(
  onActionClick: OnActionClickFn<DictionaryDictDataApi.DictData>,
  onStatusChange?: (
    newStatus: DictionaryDictDataApi.DictData['status'],
    row: DictionaryDictDataApi.DictData,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<DictionaryDictDataApi.DictData>['columns'] {
  return [
    {
      field: 'label',
      align: 'center',
      title: $t('system.dict_data.fields.label'),
    },
    {
      field: 'value',
      align: 'center',
      title: $t('system.dict_data.fields.value'),
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.dict_data.fields.status'),
      width: 140,
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('system.dict_data.fields.updatedAt'),
      width: 200,
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('system.dict_data.fields.operation'),
      width: 200,
      fixed: 'right',
      showOverflow: false,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'label',
          onClick: onActionClick,
        },
        options: ['edit', 'delete'],
      },
    },
  ];
}
