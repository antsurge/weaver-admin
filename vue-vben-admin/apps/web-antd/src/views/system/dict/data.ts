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
      field: 'id',
      type: 'checkbox',
      width: 80,
      fixed: 'left',
      align: 'center',
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

export function useDictDataColumns(
  // onActionClick: OnActionClickFn<SystemDictApi.DictData>,
  // onStatusChange?: (
  //   newStatus: SystemDictApi.DictData['status'],
  //   row: SystemDictApi.DictData,
  // ) => PromiseLike<boolean | undefined>,
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
        // attrs: { beforeChange: onStatusChange },
      },
    },
    {
      field: 'description',
      align: 'center',
      title: $t('system.dict_data.fields.description'),
      minWidth: 200,
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
          // onClick: onActionClick,
        },
        options: ['edit', 'delete'],
      },
    },
  ];
}

