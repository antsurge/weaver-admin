import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemDictApi } from '#/api/system/dict';

import { $t } from '#/locales';

export function useDictTypeColumns(
  onActionClick: OnActionClickFn<SystemDictApi.DictType>,
  onStatusChange?: (
    newStatus: SystemDictApi.DictType['status'],
    row: SystemDictApi.DictType,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemDictApi.DictType>['columns'] {
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
