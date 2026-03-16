import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemOrganizationApi } from '#/api/system/organization';

import { $t } from '#/locales';

export function useColumns(
  onActionClick: OnActionClickFn<SystemOrganizationApi.Organization>,
  onStatusChange?: (
    newStatus: SystemOrganizationApi.Organization['status'],
    row: SystemOrganizationApi.Organization,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemOrganizationApi.Organization>['columns'] {
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
      title: $t('system.organization.fields.name'),
      treeNode: true,
    },
    {
      field: 'code',
      align: 'center',
      title: $t('system.organization.fields.code'),
      width: 200,
    },
    {
      field: 'weight',
      align: 'center',
      title: $t('system.organization.fields.weight'),
      width: 120,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.organization.fields.status'),
      width: 140,
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('system.organization.fields.updatedAt'),
      width: 200,
    },
    {
      field: 'operation',
      align: 'center',
      fixed: 'right',
      showOverflow: false,
      title: $t('system.organization.fields.operation'),
      width: 260,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'name',
          onClick: onActionClick,
        },
        options: [
          {
            code: 'append',
            text: $t('system.organization.operation.appendChildren'),
          },
          'edit',
          'delete',
        ],
      },
    },
  ];
}

