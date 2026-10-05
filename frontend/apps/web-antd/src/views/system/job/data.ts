import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemJobApi } from '#/api/system/job';

import { useAccess } from '@vben/access';

import { SystemAuthCode } from '#/access';
import { $t } from '#/locales';

export function useJobColumns(
  onActionClick: OnActionClickFn<SystemJobApi.Job>,
  onStatusChange?: (
    newStatus: SystemJobApi.Job['status'],
    row: SystemJobApi.Job,
  ) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions<SystemJobApi.Job>['columns'] {
  const { hasAccessByCodes } = useAccess();
  return [
    {
      type: 'checkbox',
      width: 60,
      align: 'center',
    },
    {
      field: 'name',
      align: 'center',
      title: $t('system.job.fields.name'),
      minWidth: 140,
    },
    {
      field: 'jobGroup',
      align: 'center',
      title: $t('system.job.fields.jobGroup'),
      width: 110,
    },
    {
      field: 'jobType',
      align: 'center',
      title: $t('system.job.fields.jobType'),
      width: 90,
      slots: { default: 'job_type' },
    },
    {
      field: 'invokeTarget',
      align: 'center',
      title: $t('system.job.fields.invokeTarget'),
      minWidth: 220,
      showOverflow: true,
    },
    {
      field: 'cronExpression',
      align: 'center',
      title: $t('system.job.fields.cronExpression'),
      width: 150,
    },
    {
      field: 'misfirePolicy',
      align: 'center',
      title: $t('system.job.fields.misfirePolicy'),
      width: 110,
      slots: { default: 'misfire_policy' },
    },
    {
      field: 'concurrent',
      align: 'center',
      title: $t('system.job.fields.concurrent'),
      width: 100,
      formatter: ({ row }: { row: SystemJobApi.Job }) =>
        row.concurrent ? $t('common.yes') : $t('common.no'),
    },
    {
      field: 'nextRunTime',
      align: 'center',
      title: $t('system.job.fields.nextRunTime'),
      width: 190,
      formatter: 'formatDateTime',
    },
    {
      field: 'status',
      align: 'center',
      title: $t('system.job.fields.status'),
      width: 100,
      cellRender: {
        name: 'CellSwitch',
        attrs: {
          beforeChange: onStatusChange,
          // 无启用/禁用按钮权限时禁用开关
          disabled: () => !hasAccessByCodes([SystemAuthCode.Job.SwitchStatus]),
        },
      },
    },
    {
      field: 'updatedAt',
      align: 'center',
      title: $t('system.job.fields.updatedAt'),
      width: 160,
      formatter: 'formatDateTime',
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('system.job.fields.operation'),
      width: 220,
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
            code: 'edit',
            // 无编辑权限时不展示
            show: hasAccessByCodes([SystemAuthCode.Job.Edit]),
          },
          {
            code: 'run',
            // 立即执行：菜单未配置独立按钮权限，受后端接口权限控制
            show: hasAccessByCodes([SystemAuthCode.Job.Edit]),
          },
          {
            code: 'delete',
            // 无删除权限时不展示
            show: hasAccessByCodes([SystemAuthCode.Job.Delete]),
          },
        ],
      },
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
          placeholder: $t('system.job.search_placeholder.name'),
        },
        fieldName: 'name',
        label: $t('system.job.fields.name'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('system.job.search_placeholder.jobGroup'),
        },
        fieldName: 'jobGroup',
        label: $t('system.job.fields.jobGroup'),
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('system.job.fields.status'),
        componentProps: {
          placeholder: $t('ui.formRules.selectRequired', [
            $t('system.job.fields.status'),
          ]),
          options: [
            { label: $t('common.enabled'), value: 'enabled' },
            { label: $t('common.disabled'), value: 'disabled' },
          ],
          allowClear: true,
        },
      },
    ],
    showCollapseButton: true,
    submitButtonOptions: {
      content: $t('common.actions.search'),
    },
    submitOnChange: false,
    submitOnEnter: false,
  };
}
