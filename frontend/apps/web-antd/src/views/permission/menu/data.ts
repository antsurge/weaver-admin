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

export const PermissionBadgeTypeOptionsValueDot = "dot"
export const PermissionBadgeTypeOptionsValueText = "text"
export function getBadgeTypeOptions() {
  return [
    {
      label: $t('permission.menu.badgeType_options.dot'),
      value: PermissionBadgeTypeOptionsValueDot,
    },
    {
      color: 'success',
      label: $t('permission.menu.badgeType_options.text'),
      value: PermissionBadgeTypeOptionsValueText,
    },
  ];
}

export function getBadgeVariantsOptions() {
  return [
    {
      label: 'default',
      value: 'default',
    },
    {
      label: 'destructive',
      value: 'destructive',
    },
    {
      label: 'primary',
      value: 'primary',
    },
    {
      label: 'success',
      value: 'success',
    },
    {
      label: 'warning',
      value: 'warning',
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
      width: 60,
      fixed: 'left'
    },
    {
      field: 'title',
      align: 'center',
      title: $t('permission.menu.fields.title'),
      treeNode: true,
      slots: { default: 'title' },
      width: 200,
      formatter: ({ cellValue }) => {
        return $t(cellValue);
      },
    },
    {
      field: 'name',
      align: 'center',
      title: $t('permission.menu.fields.name'),
      width: 200,
    },
    {
      field: 'authCode',
      align: 'center',
      title: $t('permission.menu.fields.authCode'),
      width: 200,
    },
    {
      align: 'center',
      cellRender: { name: 'CellTag', options: getPermissionTypeOptions() },
      field: 'type',
      title: $t('permission.menu.fields.type'),
      width: 80,
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
      width: 100,
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
