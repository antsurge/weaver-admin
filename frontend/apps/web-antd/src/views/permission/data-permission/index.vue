<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemDataPermissionApi } from '#/api/permission/data-permission';

import { nextTick, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { IconifyIcon, Plus } from '@vben/icons';
import { $t } from '@vben/locales';

import { Button, message, Modal, Tag } from 'ant-design-vue';

import { PermissionAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteDataPermissionApi,
  getDataPermissionListApi,
  updateDataPermissionStatusApi,
} from '#/api/permission/data-permission';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import {
  formatScopeType,
  useDataPermissionColumns,
  useFormOptions,
} from './data';
import DataPermissionForm from './modules/form/index.vue';

const [DataPermissionFormModal, dataPermissionFormApi] = useVbenModal({
  connectedComponent: DataPermissionForm,
  destroyOnClose: true,
});

const selectedRows = ref<SystemDataPermissionApi.DataPermission[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useDataPermissionColumns(onActionClick, onStatusChange),
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
          const res = await getDataPermissionListApi({
            ...page,
            ...formValues,
          });
          return res;
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

  gridEvents: {
    checkboxChange() {
      nextTick(() => {
        const records = (gridApi.grid as any)?.getCheckboxRecords?.() ?? [];
        selectedRows.value = records;
      });
    },
  },
});

function onRefresh() {
  gridApi.query();
}

function onCreate() {
  dataPermissionFormApi.setData({}).open();
}

function onActionClick({
  code,
  row,
}: {
  code: string;
  row: SystemDataPermissionApi.DataPermission;
}) {
  switch (code) {
    case 'delete': {
      onDelete([row]);
      break;
    }
    case 'edit': {
      dataPermissionFormApi.setData(row).open();
      break;
    }
  }
}

function onStatusChange(
  newStatus: SystemDataPermissionApi.DataPermission['status'],
  row: SystemDataPermissionApi.DataPermission,
): Promise<boolean | undefined> {
  const statusText: Record<string, string> = {
    enabled: $t('common.enabled'),
    disabled: $t('common.disabled'),
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: $t('ui.actionTitle.switchStatus'),
      content: `${$t('permission.data_permission.message.switchStatusConfirm', [
        row.name,
        statusText[newStatus],
      ])}`,
      async onOk() {
        try {
          await updateDataPermissionStatusApi(row.id, newStatus);
          resolve(true);
        } catch {
          resolve(false);
        }
      },
      onCancel() {
        resolve(false);
      },
    });
  });
}

async function onDelete(rows?: SystemDataPermissionApi.DataPermission[]) {
  const list = rows?.length ? rows : selectedRows.value;
  if (list.length === 0) return;

  const ids = list.map((item) => item.id);
  const names = list.map((item) => item.name).join('、');

  const confirmed = await new Promise<boolean>((resolve) => {
    Modal.confirm({
      title: $t('ui.actionTitle.delete'),
      content: $t('ui.actionMessage.deleteConfirm', [names]),
      async onOk() {
        resolve(true);
      },
      onCancel() {
        resolve(false);
      },
    });
  });
  if (!confirmed) return;

  try {
    await deleteDataPermissionApi(ids);
    message.success($t('ui.actionMessage.deleteSuccess', [names]));
    selectedRows.value = [];
    onRefresh();
  } catch {
    /* 错误由全局拦截器提示 */
  }
}

/** 批量删除（工具栏按钮） */
function onBatchDelete() {
  const rows = selectedRows.value;
  if (!rows?.length) return;
  onDelete(rows);
}
</script>

<template>
  <Page auto-content-height>
    <DataPermissionFormModal @success="onRefresh" />

    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            v-access:code="[PermissionAuthCode.DataPermission.BatchDelete]"
            type="primary"
            danger
            :disabled="selectedRows.length === 0"
            @click="onBatchDelete"
          >
            <IconifyIcon icon="ant-design:delete-outlined" class="size-5" />
            {{ $t('ui.actionTitle.delete') }}
          </Button>
          <Button
            v-access:code="[PermissionAuthCode.DataPermission.Create]"
            type="primary"
            @click="onCreate"
          >
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>
        </div>
      </template>

      <template #scope_type="{ row }">
        <Tag color="blue">{{ formatScopeType(row.scopeType) }}</Tag>
      </template>
    </Grid>
  </Page>
</template>
