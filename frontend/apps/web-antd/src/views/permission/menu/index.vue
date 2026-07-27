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

import { MenuBadge } from '@vben-core/menu-ui';

import { Button, message, Modal } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteMenuApi,
  getMenuTreeApi,
  updateMenuStatusApi,
} from '#/api/permission/menu';
import type {
  PermissionMenuApi
} from '#/api/permission/menu';

import { PermissionTypeOptionsValueAction, useColumns, useFormOptions } from './data';
import Form from './modules/form/index.vue';

const [FormModel, formModelApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

const selectedRows = ref<PermissionMenuApi.PermissionMenu[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useColumns(onActionClick, onStatusChange),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: false,
    },
    proxyConfig: {
      autoLoad: true,
      ajax: {
        query: async ({ }, formValues) => {
          let res = await getMenuTreeApi(formValues);
          return res?.items ?? []
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    checkboxConfig: {
      checkStrictly: true,
      showHeader: true,
    },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
      search: true,
    },
    treeConfig: {
      parentField: 'parentID',
      rowField: 'id',
      transform: false,
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
  selectedRows.value = checkboxRecords
}

function onActionClick({
  code,
  row,
}: OnActionClickParams<PermissionMenuApi.PermissionMenu>) {
  switch (code) {
    case 'append': {
      onAppend(row);
      break;
    }
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

function onEdit(row: PermissionMenuApi.PermissionMenu) {
  formModelApi.setData(row).open();
}

function onCreate() {
  formModelApi.setData({}).open();
}
function onAppend(row: PermissionMenuApi.PermissionMenu) {
  formModelApi.setData({ parentID: row.id }).open();
}

async function onStatusChange(
  newStatus: number,
  row: PermissionMenuApi.PermissionMenu,
) {
  const status: Recordable<string> = {
    "disabled": '禁用',
    "enabled": '启用',
  };

  return new Promise((resolve) => {
    Modal.confirm({
      title: '切换状态',
      content: `你要将 ${row.name} 的状态切换为 【${status[newStatus]}】 吗？`,
      okText: '确认',
      cancelText: '取消',
      async onOk() {
        try {
          await updateMenuStatusApi(row.id, newStatus);
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

function onDelete(row: PermissionMenuApi.PermissionMenu) {
  const hideLoading = message.loading({
    content: $t('ui.actionMessage.deleting', [row.name]),
    duration: 0,
    key: 'action_process_msg',
  });

  deleteMenuApi([row.id])
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

// 展开全部
function onExpandAll() {
  gridApi.grid?.setAllTreeExpand(true);
}

// 折叠全部
function onCollapseAll() {
  gridApi.grid?.clearTreeExpand();
}

// 批量删除
function onBatchDelete() {
  const rows = selectedRows.value;
  if (!rows?.length) return;
  const ids = rows.map((r) => r.id);
  const names = rows.map((r) => r.name ?? r.meta?.title ?? r.id).join('、');

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
      return deleteMenuApi(ids)
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
          <Button type="primary" class="inline-flex items-center" @click="onExpandAll">
            <IconifyIcon icon="ant-design:column-height-outlined" class="size-5" />
            {{ $t('permission.menu.actionTitle.expandAll') }}
          </Button>
          <Button type="primary" class="inline-flex items-center" @click="onCollapseAll">
            <IconifyIcon icon="ant-design:column-width-outlined" class="size-5" />
            {{ $t('permission.menu.actionTitle.collapseAll') }}
          </Button>
          <Button type="primary" class="inline-flex items-center" danger @click="onBatchDelete"
            :disabled="!selectedRows.length">
            <IconifyIcon icon="ant-design:delete-outlined" class="size-5" />
            {{ $t('ui.actionTitle.delete') }}
          </Button>
          <Button type="primary" class="inline-flex items-center" @click="onCreate">
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>
        </div>
      </template>
      <template #title="{ row }">
        <div class="flex w-full items-center justify-between">
          <!-- 左侧 icon + title -->
          <div class="flex items-center gap-2">
            <div class="size-5 shrink-0">
              <IconifyIcon v-if="row.type === PermissionTypeOptionsValueAction" icon="carbon:security"
                class="size-full" />
              <IconifyIcon v-else-if="row?.icon" :icon="row.icon || 'carbon:circle-dash'" class="size-full" />
            </div>
            <span>
              {{ $t(row?.title) }}
            </span>
          </div>
          <!-- 右侧 Badge -->
          <MenuBadge v-if="row?.badgeType" class="menu-badge" :badge="row.badgeContent" :badge-type="row.badgeType"
            :badge-variants="row.badgeStyle" />
        </div>
      </template>
    </Grid>
  </Page>
</template>
<style lang="scss" scoped>
.menu-badge {
  top: 50%;
  right: 0;
  transform: translateY(-50%);

  &> :deep(div) {
    padding-top: 0;
    padding-bottom: 0;
  }
}
</style>
