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
  deleteRoleApi,
  getRoleListApi,
  updateRoleStatusApi,
} from '#/api/permission/role';

import type { PermissionRoleApi } from '#/api/permission/role';

import { useFormOptions, useColumns } from './data';
import Form from './modules/form/index.vue';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from "#/types/pagination"

const [FormModel, formModelApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

const selectedRows = ref<PermissionRoleApi.Role[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useColumns(onActionClick, onStatusChange),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: true,
      pageSize: DEFAULT_PAGE_SIZE,
      pageSizes: PAGE_SIZES,
    },
    proxyConfig: {
      autoLoad: true,
      ajax: {
        query: async ({ page }, formValues) => {
          const res = await getRoleListApi({
            ...page,
            ...formValues,
          });
          return res
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    checkboxConfig: {
      reserve: true,
      showReserveStatus: true,
    },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
      search: true,
    },
  } as VxeTableGridOptions,
  gridEvents: {
    checkboxChange() {
      nextTick(() => {
        onCheckboxChange()
      })
    },
    checkboxAll() {
      nextTick(() => {
        onCheckboxChange()
      })
    }
  },
});

function onCheckboxChange() {
  const checkboxRecords = gridApi.grid?.getCheckboxRecords?.() ?? []
  const checkboxReserveRecords = gridApi.grid?.getCheckboxReserveRecords?.() ?? []
  selectedRows.value = [...checkboxRecords, ...checkboxReserveRecords]
}

function onActionClick({
  code,
  row,
}: OnActionClickParams<PermissionRoleApi.Role>) {
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

function onEdit(row: PermissionRoleApi.Role) {
  formModelApi.setData(row).open();
}

function onCreate() {
  formModelApi.setData({}).open();
}

function onStatusChange(
  newStatus: PermissionRoleApi.Role['status'],
  row: PermissionRoleApi.Role,
): Promise<boolean | undefined> {
  const statusText: Recordable<string> = {
    disabled: $t('common.disabled'),
    enabled: $t('common.enabled'),
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: $t('permission.role.actions.switchStatus'),
      content: $t('permission.role.actions.switchStatusConfirm', [
        row.name,
        statusText[newStatus],
      ]),
      okText: $t('ui.actionTitle.confirm'),
      cancelText: $t('ui.actionTitle.cancel'),
      async onOk() {
        try {
          await updateRoleStatusApi(row.id, newStatus);
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

function onDelete(row: PermissionRoleApi.Role) {
  deleteRoleApi([row.id])
    .then(() => {
      onRefresh();
    })
    .catch(() => {
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
      return deleteRoleApi(ids)
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
          <Button type="primary" danger :disabled="!selectedRows.length" @click="onBatchDelete">
            <IconifyIcon icon="ant-design:delete-outlined" class="size-5" />
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
