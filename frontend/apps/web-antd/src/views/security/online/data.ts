import type { VbenFormProps } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { OnlineApi } from '#/api/security/online';

import { $t } from '#/locales';

/** Unix 秒 -> 本地时间字符串 */
function formatUnix(sec?: number) {
  if (sec === null || sec === undefined || sec <= 0) return '-';
  return new Date(sec * 1000).toLocaleString('zh-CN', { hour12: false });
}

export function useOnlineColumns(
  onActionClick: OnActionClickFn<OnlineApi.OnlineUser>,
  canForceLogout = true,
): VxeTableGridOptions<OnlineApi.OnlineUser>['columns'] {
  return [
    {
      field: 'username',
      align: 'center',
      title: $t('security.online.fields.username'),
      width: 160,
      showOverflow: 'tooltip',
    },
    {
      field: 'realName',
      align: 'center',
      title: $t('security.online.fields.realName'),
      width: 120,
      showOverflow: 'tooltip',
    },
    {
      field: 'ip',
      align: 'center',
      title: $t('security.online.fields.ip'),
      width: 140,
    },
    {
      field: 'userAgent',
      align: 'center',
      title: $t('security.online.fields.userAgent'),
      minWidth: 220,
      showOverflow: 'tooltip',
    },
    {
      field: 'sessionCount',
      align: 'center',
      title: $t('security.online.fields.sessionCount'),
      width: 100,
      formatter: ({ cellValue }) => (cellValue ? `${cellValue}` : '-'),
    },
    {
      field: 'loginAt',
      align: 'center',
      title: $t('security.online.fields.loginAt'),
      width: 180,
      formatter: ({ cellValue }) => formatUnix(cellValue),
    },
    {
      field: 'expireAt',
      align: 'center',
      title: $t('security.online.fields.expireAt'),
      width: 180,
      formatter: ({ cellValue }) => formatUnix(cellValue),
    },
    {
      field: 'operation',
      align: 'center',
      title: $t('security.online.fields.operation'),
      width: 110,
      fixed: 'right',
      showOverflow: false,
      cellRender: {
        name: 'CellOperation',
        attrs: {
          nameField: 'username',
          onClick: onActionClick,
        },
        options: [
          {
            code: 'logout',
            text: $t('security.online.operation.logout'),
            show: canForceLogout,
          },
        ],
      },
    },
  ];
}

/** 在线用户搜索表单 */
export function useOnlineFormOptions(): VbenFormProps {
  return {
    collapsed: false,
    schema: [
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('security.online.search_placeholder.username'),
        },
        fieldName: 'username',
        label: $t('security.online.fields.username'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('security.online.search_placeholder.realName'),
        },
        fieldName: 'realName',
        label: $t('security.online.fields.realName'),
      },
      {
        component: 'Input',
        componentProps: {
          placeholder: $t('security.online.search_placeholder.ip'),
        },
        fieldName: 'ip',
        label: $t('security.online.fields.ip'),
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
