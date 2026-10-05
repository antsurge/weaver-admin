import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { PermissionMenuApi } from '#/api/permission/menu';

import { useAccess } from '@vben/access';

import { PermissionAuthCode } from '#/access';
import { $t } from '#/locales';

export const PermissionTypeOptionsValueCatalog = 'catalog';
export const PermissionTypeOptionsValueMenu = 'menu';
export const PermissionTypeOptionsValueAction = 'action';
export const PermissionTypeOptionsValueIframe = 'iframe';
export const PermissionTypeOptionsValueLink = 'link';
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

export const PermissionBadgeTypeOptionsValueDot = 'dot';
export const PermissionBadgeTypeOptionsValueText = 'text';
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

// 标题、名称、权限标识、类型、路径、权重、状态、修改时间
export function useColumns(
  onActionClick: OnActionClickFn<PermissionMenuApi.PermissionMenu>,
  onStatusChange?: (
    newStatus: any,
    row: PermissionMenuApi.PermissionMenu,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<PermissionMenuApi.PermissionMenu>['columns'] {
  const { hasAccessByCodes } = useAccess();
  return [
    {
      field: 'id',
      align: 'center',
      type: 'checkbox',
      width: 60,
      fixed: 'left',
    },
    {
      field: 'title',
      align: 'center',
      title: $t('permission.menu.fields.title'),
      treeNode: true,
      slots: { default: 'title' },
      width: 160,
      formatter: ({ cellValue }) => {
        return $t(cellValue);
      },
    },
    {
      field: 'name',
      align: 'center',
      title: $t('permission.menu.fields.name'),
      width: 160,
    },
    {
      field: 'authCode',
      align: 'center',
      title: $t('permission.menu.fields.authCode'),
      width: 220,
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
      align: 'center',
      field: 'weight',
      title: $t('permission.menu.fields.weight'),
      width: 80,
    },
    {
      field: 'status',
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无启用/禁用按钮权限时禁用开关
          disabled: () =>
            !hasAccessByCodes([PermissionAuthCode.Menu.SwitchStatus]),
        },
      },
      align: 'center',
      title: $t('permission.menu.fields.status'),
      width: 100,
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('permission.menu.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      align: 'center',
      cellRender: {
        attrs: {
          nameField: 'name',
          onClick: onActionClick,
        },
        name: 'CellOperation',
        options: [
          {
            code: 'append',
            text: $t('permission.menu.operation.appendChildren'),
            // 追加子级属于独立按钮权限
            show: (row: any) =>
              row.type !== PermissionTypeOptionsValueAction &&
              hasAccessByCodes([PermissionAuthCode.Menu.AddSubordinate]),
          },
          {
            code: 'edit',
            show: hasAccessByCodes([PermissionAuthCode.Menu.Edit]),
          },
          {
            code: 'delete',
            show: hasAccessByCodes([PermissionAuthCode.Menu.Delete]),
          },
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
        fieldName: 'type',
        label: $t('permission.menu.fields.type'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('permission.menu.fields.type'),
          ]),
          options: getPermissionTypeOptions(),
          allowClear: true,
        },
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
            { label: $t('common.enabled'), value: 'enabled' },
            { label: $t('common.disabled'), value: 'disabled' },
          ],
          allowClear: true,
        },
      },
    ],
    // 控制表单是否显示折叠按钮
    showCollapseButton: true,
    submitButtonOptions: {
      content: $t('common.actions.search'),
    },
    // 是否在字段值改变时提交表单
    submitOnChange: false,
    // 按下回车时是否提交表单
    submitOnEnter: false,
  };
}
