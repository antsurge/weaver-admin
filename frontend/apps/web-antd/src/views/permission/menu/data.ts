import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { PermissionMenuApi } from '#/api/permission/menu';
import type { VbenFormProps } from '#/adapter/form';

import { $t } from '#/locales';

export const PermissionTypeOptionsValueCatalog = "catalog"
export const PermissionTypeOptionsValueMenu = "menu"
export const PermissionTypeOptionsValueAction = "action"
export const PermissionTypeOptionsValueIframe = "iframe"
export const PermissionTypeOptionsValueLink = "link"
export function getPermissionTypeOptions() {
  return [
    {
      color: 'purple',
      label: $t('permission.menu.type_options.catalog'),
      value: 'catalog',
    },
    {
      color: 'success',
      label: $t('permission.menu.type_options.menu'),
      value: 'menu',
    },
    {
      color: 'warning',
      label: $t('permission.menu.type_options.action'),
      value: 'action',
    },
    {
      color: 'blue',
      label: $t('permission.menu.type_options.iframe'),
      value: 'iframe',
    },
    {
      color: 'cyan',
      label: $t('permission.menu.type_options.link'),
      value: 'link',
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
      label: $t('permission.menu.menu_type_options.tab'),
      value: PermissionMenuTypeOptionsValueTab,
    },
    {
      color: 'success',
      label: $t('permission.menu.menu_type_options.link'),
      value: PermissionMenuTypeOptionsValueLink,
    },
    {
      color: 'warning',
      label: $t('permission.menu.menu_type_options.iframe'),
      value: PermissionMenuTypeOptionsValueIframe,
    },
  ];
}

// 标题，图标，Code、类型、状态、修改时间
export function useColumns(
  onActionClick: OnActionClickFn<PermissionMenuApi.PermissionMenu>,
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<PermissionMenuApi.PermissionMenu>['columns'] {
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
      title: $t('permission.menu.fields.name'),
      treeNode: true,
      width: 250,
    },
    {
      field: 'icon',
      align: 'center',
      title: $t('permission.menu.fields.icon'),
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
      title: $t('permission.menu.fields.code'),
      width: 200,
    },
    {
      align: 'center',
      cellRender: { name: 'CellTag', options: getPermissionTypeOptions() },
      field: 'type',
      title: $t('permission.menu.fields.type'),
      width: 160,
    },
    {
      align: 'center',
      field: 'path',
      title: $t('permission.menu.fields.path'),
    },
    {
      field: 'status',
      cellRender: {
        name: 'CellSwitch',
        attrs: { beforeChange: onStatusChange },
      },
      align: 'center',
      title: $t('permission.menu.fields.status'),
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
            text: $t('permission.menu.operation.appendChildren'),
            show: (row: any) => row.type !== PermissionTypeOptionsValueAction,
          },
          'edit', // 默认的编辑按钮
          'delete', // 默认的删除按钮
        ],
      },
      field: 'operation',
      fixed: 'right',
      showOverflow: false,
      title: $t('permission.menu.fields.operation'),
      width: 200,
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
            $t('permission.menu.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('permission.menu.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('permission.menu.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('permission.menu.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('permission.menu.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('permission.menu.fields.status'),
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
