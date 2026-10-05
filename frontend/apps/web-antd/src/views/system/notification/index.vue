<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { NotificationApi } from '#/api/system/notification';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';
import { $t } from '@vben/locales';

import { Button, message, Modal } from 'ant-design-vue';

import { SystemAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  getNotificationManageListApi,
  revokeNotificationApi,
} from '#/api/system/notification';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import { useFormOptions, useNotificationColumns } from './data';
import NotificationForm from './modules/form/index.vue';

const [NotificationFormModal, notificationFormApi] = useVbenModal({
  connectedComponent: NotificationForm,
  destroyOnClose: true,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useNotificationColumns(onActionClick),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: true,
      pageSize: DEFAULT_PAGE_SIZE,
      pageSizes: PAGE_SIZES,
    },
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getNotificationManageListApi({
            ...page,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
      search: true,
    },
  } as VxeTableGridOptions,
});

function onRefresh() {
  gridApi.query();
}

function onCreate() {
  notificationFormApi.setData({}).open();
}

function onActionClick({
  code,
  row,
}: {
  code: string;
  row: NotificationApi.Notification;
}) {
  if (code === 'revoke') {
    onRevoke(row);
  }
}

function onRevoke(row: NotificationApi.Notification) {
  if (row.status === 'revoked') {
    message.warning($t('system.notification.message.alreadyRevoked'));
    return;
  }

  Modal.confirm({
    title: $t('system.notification.confirm.revokeTitle'),
    content: $t('system.notification.confirm.revokeContent'),
    async onOk() {
      await revokeNotificationApi(row.id);
      message.success($t('system.notification.message.revokeSuccess'));
      onRefresh();
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <NotificationFormModal @success="onRefresh" />

    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            v-access:code="[SystemAuthCode.Notification.Create]"
            type="primary"
            @click="onCreate"
          >
            <Plus class="size-5" />
            {{ $t('system.notification.operation.publish') }}
          </Button>
        </div>
      </template>
    </Grid>
  </Page>
</template>
