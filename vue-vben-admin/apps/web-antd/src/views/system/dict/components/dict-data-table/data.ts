import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemDictApi } from '#/api/system/dict';
import { $t } from '#/locales';
export function useDictDataColumns(
  onActionClick: OnActionClickFn<SystemDictApi.DictData>,
  onStatusChange?: (
    newStatus: SystemDictApi.DictData['status'],
    row: SystemDictApi.DictData,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemDictApi.DictData>['columns'] {
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
