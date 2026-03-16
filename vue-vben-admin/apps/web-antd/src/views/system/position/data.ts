import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemPositionApi } from '#/api/system/position';

import { $t } from '#/locales';

export function useColumns(
  onActionClick: OnActionClickFn<SystemPositionApi.Position>,
  onStatusChange?: (
    newStatus: SystemPositionApi.Position['status'],
    row: SystemPositionApi.Position,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemPositionApi.Position>['columns'] {
  return [
    {
      field: 'id',
      align: 'center',
      type: 'checkbox',
      width: 80,
      fixed: 'left',
    },
    {
      field: 'name',
      align: 'center',
      title: $t('system.position.fields.name'),
      width: 200,
    },
    {
      field: 'code',
      align: 'center',
      title: $t('system.position.fields.code'),
      width: 200,
    },
    {
      field: 'weight',
      align: 'center',
      title: $t('system.position.fields.weight'),
      width: 120,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.position.fields.status'),
      width: 140,
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
    },
    {
      field: 'operation',
      align: 'center',
      fixed: 'right',
      showOverflow: false,
      title: $t('system.position.fields.operation'),
      width: 200,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'name',
          onClick: onActionClick,
        },
        options: ['edit', 'delete'],
      },
    },
  ];
}
