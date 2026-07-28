<script lang="ts" setup>
// ==================== types ====================
import type { OnActionClickParams } from '#/adapter/vxe-table';
import type { SystemApiInterfaceApi } from '#/api/system/api-interface';

// ==================== vue ====================
import { nextTick, ref } from 'vue';

// ==================== vben ====================
import { Page } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { $t } from '@vben/locales';

// ==================== third-party ====================
import { Button, message, Modal, Upload } from 'ant-design-vue';

// ==================== constants ====================
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

// ==================== business ====================
import { useGridFormOptions, useColumns } from './data';

// ==================== api ====================
import {
  deleteApiInterfaceApi,
  getApiInterfaceListApi,
  importApiInterfaceApi,
} from '#/api/system/api-interface';

const selectedRows = ref<SystemApiInterfaceApi.ApiInterface[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useGridFormOptions(),
  gridOptions: {
    columns: useColumns(onActionClick),
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
        query: async ({ page }: { page: { currentPage: number; pageSize: number } }, formValues: Record<string, any>) => {
          const res = await getApiInterfaceListApi({
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
  },
  gridEvents: {
    checkboxChange() {
      nextTick(() => {
        onCheckboxChange();
      });
    },
    checkboxAll() {
      nextTick(() => {
        onCheckboxChange();
      });
    },
  },
});

function onCheckboxChange() {
  const grid = gridApi.grid;
  if (!grid) return;

  selectedRows.value = [
    ...(grid.getCheckboxRecords?.() ?? []),
    ...(grid.getCheckboxReserveRecords?.() ?? []),
  ];
}

function onActionClick({
  code,
  row,
}: OnActionClickParams<SystemApiInterfaceApi.ApiInterface>) {
  switch (code) {
    case 'delete': {
      onDelete(row);
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

function onDelete(row: SystemApiInterfaceApi.ApiInterface) {
  deleteApiInterfaceApi([row.id])
    .then(() => {
      onRefresh();
    })
    .catch(() => {});
}

function onBatchDelete() {
  const rows = selectedRows.value;
  if (!rows?.length) return;
  const ids = rows.map((r) => r.id);
  const names = rows.map((r) => r.code || r.id).join('、');

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
      return deleteApiInterfaceApi(ids)
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

function beforeUpload(file: File) {
  const isYaml =
    file.name.endsWith('.yaml') ||
    file.name.endsWith('.yml');

  if (!isYaml) {
    message.error($t('system.api_interface.uploadOnlyYaml'));
    return Upload.LIST_IGNORE;
  }

  handleImport(file);
  return false; // 阻止默认上传
}

async function handleImport(file: File) {
  const hide = message.loading($t('system.api_interface.importing'), 0);

  try {
    const formData = new FormData();
    formData.append('file', file);

    const res = await importApiInterfaceApi(formData);
    message.success(
      $t('system.api_interface.importSuccess', [
        String(res.imported),
        String(res.skipped),
      ]),
    );
    onRefresh();
  } catch (e: any) {
    message.error(e?.message || $t('system.api_interface.importFailed'));
  } finally {
    hide();
  }
}
</script>

<template>
  <Page auto-content-height>
    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Upload
            :before-upload="beforeUpload"
            :show-upload-list="false"
            accept=".yaml,.yml"
          >
            <Button>
              <IconifyIcon
                icon="ant-design:upload-outlined"
                class="size-5"
              />
              {{ $t('system.api_interface.import') }}
            </Button>
          </Upload>

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
        </div>
      </template>
    </Grid>
  </Page>
</template>