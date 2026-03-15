import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemPermissionApi } from '#/api/system/permission';

import { $t } from '#/locales';

export const PermissionTypeOptionsValueMenuDir = "menu_dir"
export const PermissionTypeOptionsValueMenu = "menu"
export const PermissionTypeOptionsValueButton = "button"
export function getPermissionTypeOptions() {
  return [
    {
      color: 'purple',
      label: $t('system.permission.type_options.menu_dir'),
      value: 'menu_dir',
    },
    {
      color: 'success',
      label: $t('system.permission.type_options.menu'),
      value: 'menu'
    },
    {
      color: 'warning',
      label: $t('system.permission.type_options.button'),
      value: 'button'
    },
  ];
}

export const PermissionMenuTypeOptionsValueTab = "tab"
export const PermissionMenuTypeOptionsValueLink = "link"
export const PermissionMenuTypeOptionsValueIframe = "iframe"
export function getPermissionMenuTypeOptions() {
  return [
    {
      color: 'processing',
      label: $t('system.permission.menu_type_options.tab'),
      value: PermissionMenuTypeOptionsValueTab,
    },
    {
      color: 'success',
      label: $t('system.permission.menu_type_options.link'),
      value: PermissionMenuTypeOptionsValueLink,
    },
    {
      color: 'warning',
      label: $t('system.permission.menu_type_options.iframe'),
      value: PermissionMenuTypeOptionsValueIframe,
    },
  ];
}

// 标题，图标，Code、类型、状态、修改时间
export function useColumns(
  onActionClick: OnActionClickFn<SystemPermissionApi.SystemPermission>,
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemPermissionApi.SystemPermission>['columns'] {
  return [
    {
      field: 'id',
      align: 'center',
      type: 'checkbox',
      width: 100,
      fixed: 'left'
    },
    {
      field: 'name',
      align: 'center',
      title: $t('system.permission.fields.name'),
      treeNode: true,
      width: 250,
    },
    {
      field: 'icon',
      align: 'center',
      title: $t('system.permission.fields.icon'),
      width: 80,
      cellRender: {
        name: 'CellIcon', // 使用 Vben 封装的图标渲染器
        props: {
          // 如果图标字段名不是直接的 icon，可能需要映射，但这里 field 已经是 permission.icon
          // 通常 CellIcon 会自动读取当前单元格的值作为 icon name
        }
      },
    },
    {
      field: 'code',
      align: 'center',
      title: $t('system.permission.fields.code'),
      width: 200,
    },
    {
      align: 'center',
      cellRender: { name: 'CellTag', options: getPermissionTypeOptions() },
      field: 'type',
      title: $t('system.permission.fields.type'),
      width: 160,
    },
    {
      align: 'center',
      field: 'path',
      title: $t('system.permission.fields.path'),
    },
    {
      field: 'status',
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
      align: 'center',
      title: $t('system.permission.fields.status'),
      width: 160,
    },
    {
      align: 'center',
      cellRender: {
        attrs: {
          nameField: 'permission.name',
          onClick: onActionClick,
        },
        name: 'CellOperation',
        options: [
          {
            code: 'append',
            text: '新增下级',
            show: (row: any) => row.type !== PermissionTypeOptionsValueButton,
          },
          'edit', // 默认的编辑按钮
          'delete', // 默认的删除按钮
        ],
      },
      field: 'operation',
      fixed: 'right',
      showOverflow: false,
      title: $t('system.permission.fields.operation'),
      width: 200,
    },
  ];
}
