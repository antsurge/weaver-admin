import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { DictionaryDictDataApi } from '#/api/system/dictionary/dict-data';

import { useAccess } from '@vben/access';

import { SystemAuthCode } from '#/access';
import { $t } from '#/locales';

export function useDictDataColumns(
  onActionClick: OnActionClickFn<DictionaryDictDataApi.DictData>,
  onStatusChange?: (
    newStatus: DictionaryDictDataApi.DictData['status'],
    row: DictionaryDictDataApi.DictData,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<DictionaryDictDataApi.DictData>['columns'] {
  const { hasAccessByCodes } = useAccess();
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
      title: $t('system.dict_data.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
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
        options: [
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
