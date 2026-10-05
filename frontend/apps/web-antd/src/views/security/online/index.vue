<script lang="ts" setup>
import type { OnActionClickParams } from '#/adapter/vxe-table';
import type { OnlineApi } from '#/api/security/online';

import { computed } from 'vue';

import { useAccess } from '@vben/access';
import { Page } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { message, Modal } from 'ant-design-vue';

import { SecurityAuthCode } from '#/access/security';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { forceLogoutApi, getOnlineUsersApi } from '#/api/security/online';

import { useOnlineColumns, useOnlineFormOptions } from './data';

const { hasAccessByCodes } = useAccess();
/** 是否有强制下线权限 */
const canForceLogout = computed(() =>
  hasAccessByCodes([SecurityAuthCode.Online.ForceLogout]),
);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useOnlineFormOptions(),
  gridOptions: {
    columns: useOnlineColumns(onActionClick, canForceLogout.value),
    height: 'auto',
    keepSource: true,
    rowConfig: {
      keyField: 'id',
    },
    proxyConfig: {
      autoLoad: true,
      ajax: {
        query: async (_params, formValues) => {
          const res = await getOnlineUsersApi();
          const { username, realName, ip } = formValues || {};
          let items = res.items || [];
          // 前端过滤：用户名 / 姓名 / 登录IP
          if (username) {
            const kw = `${username}`.toLowerCase();
            items = items.filter((item) =>
              `${item.username || ''}`.toLowerCase().includes(kw),
            );
          }
          if (realName) {
            const kw = `${realName}`.toLowerCase();
            items = items.filter((item) =>
              `${item.realName || ''}`.toLowerCase().includes(kw),
            );
          }
          if (ip) {
            const kw = `${ip}`.toLowerCase();
            items = items.filter((item) =>
              `${item.ip || ''}`.toLowerCase().includes(kw),
            );
          }
          return { items, total: items.length };
        },
      },
    },
    toolbarConfig: {
      custom: false,
      export: false,
      refresh: true,
      zoom: true,
      search: true,
    },
  },
});

function onRefresh() {
  gridApi.query();
}

function onActionClick({
  code,
  row,
}: OnActionClickParams<OnlineApi.OnlineUser>) {
  switch (code) {
    case 'logout': {
      if (canForceLogout.value) {
        onForceLogout(row);
      }
      break;
    }
    default: {
      break;
    }
  }
}

function onForceLogout(row: OnlineApi.OnlineUser) {
  const name = row.username || row.realName || row.id;
  Modal.confirm({
    title: $t('security.online.confirm.logoutTitle'),
    content: $t('security.online.confirm.logoutContent', [name]),
    okText: $t('security.online.operation.logout'),
    okType: 'danger',
    onOk() {
      return forceLogoutApi(row.id)
        .then(() => {
          message.success($t('security.online.message.logoutSuccess'));
          onRefresh();
        })
        .catch(() => {});
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <Grid />
  </Page>
</template>
