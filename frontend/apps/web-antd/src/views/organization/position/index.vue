<script lang="ts" setup>
// ==================== types ====================
import type { Recordable } from '@vben/types';

import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { OrganizationPositionApi } from '#/api/organization/position';

// ==================== vue ====================
import { nextTick, ref } from 'vue';

// ==================== vben ====================
import { Page, useVbenModal } from '@vben/common-ui';
import { IconifyIcon, Plus } from '@vben/icons';
import { $t } from '@vben/locales';

// ==================== third-party ====================
import { Button, message, Modal, Upload } from 'ant-design-vue';

import { PermissionAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
// ==================== api ====================
import {
  deletePositionApi,
  downloadPositionTemplateApi,
  exportPositionApi,
  getPositionListApi,
  importPositionApi,
  updatePositionStatusApi,
} from '#/api/organization/position';
// ==================== constants ====================
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';
import {
  downloadFile,
  getFileNameFromDisposition,
  handleBlobResponseError,
} from '#/utils/download';

// ==================== business ====================
import { useColumns, useGridFormOptions } from './data';
// ==================== components ====================
import Form from './modules/form/index.vue';

const [FormModel, formModelApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

const selectedRows = ref<OrganizationPositionApi.Position[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useGridFormOptions(),
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
          const res = await getPositionListApi({
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
  } as VxeTableGridOptions,
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
}: OnActionClickParams<OrganizationPositionApi.Position>) {
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

function onEdit(row: OrganizationPositionApi.Position) {
  formModelApi.setData(row).open();
}

function onCreate() {
  formModelApi.setData({}).open();
}

function onStatusChange(
  newStatus: OrganizationPositionApi.Position['status'],
  row: OrganizationPositionApi.Position,
): Promise<boolean | undefined> {
  const statusText: Recordable<string> = {
    disabled: $t('common.disabled'),
    enabled: $t('common.enabled'),
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: $t('organization.position.actions.switchStatus'),
      content: $t('organization.position.actions.switchStatusConfirm', [
        row.name,
        statusText[newStatus],
      ]),
      okText: $t('ui.actionTitle.confirm'),
      cancelText: $t('ui.actionTitle.cancel'),
      async onOk() {
        try {
          await updatePositionStatusApi(row.id, newStatus);
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

function onDelete(row: OrganizationPositionApi.Position) {
  deletePositionApi([row.id])
    .then(() => {
      onRefresh();
    })
    .catch(() => {});
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

async function onDownloadTemplate() {
  try {
    const res = await downloadPositionTemplateApi();
    const blobData = res.data;
    const filename = getFileNameFromDisposition(
      res.headers['content-disposition'],
    );
    // 下载
    downloadFile(blobData, filename);
  } catch (error: any) {
    handleBlobResponseError(error);
  }
}

async function onExport() {
  try {
    // ⭐ 获取当前查询条件（重点！）
    const formValues = await gridApi.formApi?.getValues?.();

    const params = {
      ...formValues,
      currentPage: 1,
      pageSize: 0,
    };

    const res = await exportPositionApi(params);
    const blobData = res.data;
    const filename = getFileNameFromDisposition(
      res.headers['content-disposition'],
    );
    // 下载
    downloadFile(blobData, filename);
  } catch (error: any) {
    handleBlobResponseError(error);
  }
}

function beforeUpload(file: File) {
  const isExcel =
    file.type.includes('sheet') ||
    file.name.endsWith('.xlsx') ||
    file.name.endsWith('.xls');

  if (!isExcel) {
    message.error('只能上传 Excel 文件');
    return Upload.LIST_IGNORE;
  }

  handleImport(file);
  return false; // 阻止默认上传
}

async function handleImport(file: File) {
  const hide = message.loading('导入中...', 0);

  try {
    const formData = new FormData();
    formData.append('file', file);

    await importPositionApi(formData);

    message.success('导入成功');
    onRefresh();
  } catch (error: any) {
    message.error(error?.message || '导入失败');
  } finally {
    hide();
  }
}
</script>

<template>
  <Page auto-content-height>
    <FormModel @success="onRefresh" />
    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            v-access:code="[PermissionAuthCode.Position.Create]"
            type="primary"
            @click="onCreate"
          >
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>

          <Upload
            v-access:code="[PermissionAuthCode.Position.Import]"
            :before-upload="beforeUpload"
            :show-upload-list="false"
            accept=".xlsx,.xls"
          >
            <Button>
              <IconifyIcon icon="ant-design:upload-outlined" class="size-5" />
              {{ $t('ui.actionTitle.import') }}
            </Button>
          </Upload>

          <Button
            v-access:code="[PermissionAuthCode.Position.DownloadTemplate]"
            @click="onDownloadTemplate"
          >
            <IconifyIcon icon="ant-design:file-excel-outlined" class="size-5" />
            {{ $t('common.actions.downloadTemplate') }}
          </Button>

          <Button
            v-access:code="[PermissionAuthCode.Position.Export]"
            @click="onExport"
          >
            <IconifyIcon icon="ant-design:download-outlined" class="size-5" />
            {{ $t('ui.actionTitle.export') }}
          </Button>

          <Button
            v-access:code="[PermissionAuthCode.Position.BatchDelete]"
            type="primary"
            danger
            :disabled="selectedRows.length === 0"
            @click="onBatchDelete"
          >
            <IconifyIcon icon="ant-design:delete-outlined" class="size-5" />
            {{ $t('ui.actionTitle.delete') }}
          </Button>
        </div>
      </template>
    </Grid>
  </Page>
</template>
