<script lang="ts" setup>
import type { Recordable } from '@vben/types';
import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';

import { nextTick, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { IconifyIcon, Plus } from '@vben/icons';
import { $t } from '@vben/locales';

import { Button, message, Modal } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deletePositionApi,
  getPositionListApi,
  updatePositionStatusApi,
} from '#/api/system/position';
import type { SystemPositionApi } from '#/api/system/position';

import { useColumns } from './data';
import Form from './modules/form/index.vue';

const [FormModel, formModelApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

const selectedRows = ref<SystemPositionApi.Position[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: useColumns(onActionClick, onStatusChange),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: false,
    },
    proxyConfig: {
      ajax: {
        query: async (_params) => {
          const res = await getPositionListApi();
          const list = (res as { data?: SystemPositionApi.Position[] })?.data ?? res;
          return Array.isArray(list) ? list : [];
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    checkboxConfig: { range: false },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
    },
  } as VxeTableGridOptions,
  gridEvents: {
    checkboxChange() {
      nextTick(() => {
        const records =
          (gridApi.grid as any)?.getCheckboxRecords?.() ?? [];
        selectedRows.value = records;
      });
    },
  },
});

function onActionClick({
  code,
  row,
}: OnActionClickParams<SystemPositionApi.Position>) {
  switch (code) {
    case 'delete': {
      onDelete(row);
      break;
    }
    case 'edit': {
      onEdit(row);
      break;
    }
    default: {
      break;
    }
  }
}

function onRefresh() {
  gridApi.query();
}

function onEdit(row: SystemPositionApi.Position) {
  formModelApi.setData(row).open();
}

function onCreate() {
  formModelApi.setData({}).open();
}

function onStatusChange(
  newStatus: SystemPositionApi.Position['status'],
  row: SystemPositionApi.Position,
): Promise<boolean | undefined> {
  const statusText: Recordable<string> = {
    disabled: $t('common.disabled'),
    enabled: $t('common.enabled'),
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: $t('system.position.actions.switchStatus'),
      content: $t('system.position.actions.switchStatusConfirm', [
        row.name,
        statusText[newStatus],
      ]),
      okText: $t('ui.actionTitle.confirm'),
      cancelText: $t('ui.actionTitle.cancel'),
      async onOk() {
        try {
          await updatePositionStatusApi(row.id, newStatus);
          resolve(true);
        } catch (e) {
          resolve(false);
        }
      },
      onCancel() {
        resolve(false);
      },
    });
  });
}

function onDelete(row: SystemPositionApi.Position) {
  const hideLoading = message.loading({
    content: $t('ui.actionMessage.deleting', [row.name]),
    duration: 0,
    key: 'action_process_msg',
  });

  deletePositionApi([row.id])
    .then(() => {
      message.success({
        content: $t('ui.actionMessage.deleteSuccess', [row.name]),
        key: 'action_process_msg',
      });
      onRefresh();
    })
    .catch(() => {
      hideLoading();
    });
}

function onBatchDelete() {
  const rows = selectedRows.value;
  if (!rows?.length) return;
  const ids = rows.map((r) => r.id);
  const names = rows.map((r) => r.name || r.id).join('、');

  Modal.confirm({
    title: $t('ui.actionMessage.confirmDelete'),
    content: $t('ui.actionMessage.deleteConfirm', [names]),
    okText: $t('ui.actionTitle.confirm'),
    cancelText: $t('ui.actionTitle.cancel'),
    okType: 'danger',
    onOk() {
      const hideLoading = message.loading({
        content: $t('ui.actionMessage.deleting', [names]),
        duration: 0,
        key: 'batch_delete_msg',
      });
      return deletePositionApi(ids)
        .then(() => {
          message.success({
            content: $t('ui.actionMessage.deleteSuccess', [names]),
            key: 'batch_delete_msg',
          });
          selectedRows.value = [];
          onRefresh();
        })
        .catch(() => {
          hideLoading();
        });
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <FormModel @success="onRefresh" />
    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            type="primary"
            danger
            :disabled="!selectedRows.length"
            @click="onBatchDelete"
          >
            <IconifyIcon
              icon="ant-design:delete-outlined"
              class="size-5"
            />
            {{ $t('ui.actionTitle.delete') }}
          </Button>
          <Button type="primary" @click="onCreate">
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>
        </div>
      </template>
    </Grid>
  </Page>
</template>
