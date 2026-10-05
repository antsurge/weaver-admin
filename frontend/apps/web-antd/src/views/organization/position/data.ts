import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { OrganizationPositionApi } from '#/api/organization/position';

import { useAccess } from '@vben/access';

import { PermissionAuthCode } from '#/access';
import { z } from '#/adapter/form';
import { isPositionNameExistsApi } from '#/api/organization/position';
import { $t } from '#/locales';

export const rules = {
  /**
   * 职务名称
   */
  nameRule: z
    .string()
    .min(
      2,
      $t('ui.formRules.minLength', [
        $t('organization.position.fields.name'),
        2,
      ]),
    )
    .max(
      30,
      $t('ui.formRules.maxLength', [
        $t('organization.position.fields.name'),
        30,
      ]),
    )
    .refine(
      async (value: string) => {
        const res = await isPositionNameExistsApi(value);
        return !res?.exists;
      },
      (value) => ({
        message: $t('ui.formRules.alreadyExists', [
          $t('organization.position.fields.name'),
          value,
        ]),
      }),
    ),

  /**
   * 职务编码
   */
  codeRule: z
    .string()
    .min(
      2,
      $t('ui.formRules.minLength', [
        $t('organization.position.fields.code'),
        2,
      ]),
    )
    .max(
      30,
      $t('ui.formRules.maxLength', [
        $t('organization.position.fields.code'),
        30,
      ]),
    ),
};

export function useColumns(
  onActionClick: OnActionClickFn<OrganizationPositionApi.Position>,
  onStatusChange?: (
    newStatus: OrganizationPositionApi.Position['status'],
    row: OrganizationPositionApi.Position,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<OrganizationPositionApi.Position>['columns'] {
  const { hasAccessByCodes } = useAccess();

  return [
    {
      type: 'checkbox',
      align: 'center',
      width: 60,
      fixed: 'left',
    },
    {
      type: 'seq',
      title: $t('common.fields.seq'),
      width: 60,
      align: 'center',
    },
    {
      field: 'name',
      align: 'center',
      title: $t('organization.position.fields.name'),
    },
    {
      field: 'code',
      align: 'center',
      title: $t('organization.position.fields.code'),
      width: 260,
    },
    {
      field: 'weight',
      align: 'center',
      title: $t('organization.position.fields.weight'),
      width: 80,
    },
    {
      field: 'status',
      align: 'center',
      title: $t('organization.position.fields.status'),
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无状态切换按钮权限时禁用开关
          disabled: () => !hasAccessByCodes([PermissionAuthCode.Position.Edit]),
        },
      },
    },
    {
      field: 'createdAt',
      align: 'center',
      title: $t('common.fields.createdAt'),
      width: 160,
      formatter: ({ cellValue }) =>
        cellValue ? new Date(cellValue).toLocaleString() : '-', // 格式化
    },
    {
      field: 'operation',
      align: 'center',
      fixed: 'right',
      showOverflow: false,
      title: $t('common.fields.operation'),
      width: 200,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'name',
          onClick: onActionClick,
        },
        options: [
          {
            code: 'edit',
            // 无编辑按钮权限时不展示
            show: hasAccessByCodes([PermissionAuthCode.Position.Edit]),
          },
          {
            code: 'delete',
            // 无删除按钮权限时不展示
            show: hasAccessByCodes([PermissionAuthCode.Position.Delete]),
          },
        ],
      },
    },
  ];
}

export function useGridFormOptions(): VbenFormProps {
  return {
    collapsed: false,
    schema: [
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('organization.position.fields.name'),
          ]),
        },
        fieldName: 'name',
        label: $t('organization.position.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('ui.formRules.required', [
            $t('organization.position.fields.code'),
          ]),
        },
        fieldName: 'code',
        label: $t('organization.position.fields.code'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('organization.position.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('organization.position.fields.status'),
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
  };
}
