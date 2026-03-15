<script lang="ts" setup>
import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';

import {ref} from "vue"

import { Page, useVbenModal } from '@vben/common-ui';
import { IconifyIcon, Plus } from '@vben/icons';
import { $t } from '@vben/locales';

// import { MenuBadge } from '@vben-core/menu-ui';

import { Button, message } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { deletePermission, getPermissionTreeApi, SystemPermissionApi } from '#/api/system/permission';

import { useColumns } from './data';
import Form from './modules/form/index.vue';

const [FormModel, formModelApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

const selectedRows = ref<SystemPermissionApi.SystemPermission[]>([]);

function onSelectionChange(rows: SystemPermissionApi.SystemPermission[]) {
  selectedRows.value = rows;
}

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: useColumns(onActionClick),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: false,
    },
    proxyConfig: {
      ajax: {
        query: async (_params) => {
          let res = await getPermissionTreeApi();
          return res?.data ?? []
        },
      },
    },
    rowConfig: {
      keyField: 'permission.id',
    },
    checkboxConfig: { range: false }, // 开启多选
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
    },
    treeConfig: {
      parentField: 'pid',
      rowField: 'id',
      transform: false,
    },
  } as VxeTableGridOptions,
});

function onActionClick({
  code,
  row,
}: OnActionClickParams<SystemPermissionApi.SystemPermission>) {
  switch (code) {
    case 'append': {
      onAppend(row);
      break;
    }
    case 'delete': {
      onDelete([row]);
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

function onEdit(row: SystemPermissionApi.SystemPermission) {
  formModelApi.setData(row).open();
}

function onCreate() {
  formModelApi.setData({}).open();
}
function onAppend(row: SystemPermissionApi.SystemPermission) {
  formModelApi.setData({ parentId: row.id }).open();
}

function onDelete(row: SystemPermissionApi.SystemPermission) {
  const hideLoading = message.loading({
    content: $t('ui.actionMessage.deleting',[row.name]),
    duration: 0,
    key: 'action_process_msg',
  });

  deletePermission([row.id])
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

// 批量删除
function onBatchDelete() {
  console.log("批量删除")
}

</script>
<template>
  <Page auto-content-height>
    <FormModel @success="onRefresh" />
    <Grid @checkbox-change="onSelectionChange">
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button type="primary" @click="onBatchDelete" :disabled="selectedRows.length === 0">
            <IconifyIcon icon="ant-design:delete-outlined" class="size-5" />
            {{ $t('ui.actionTitle.delete') }}
          </Button>
          <Button type="primary" @click="onCreate">
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>
        </div>
      </template>
      <template #title="{ row }">
        <div class="flex w-full items-center gap-1">
          <div class="size-5 shrink-0">
            <IconifyIcon v-if="row.type === 'button'" icon="carbon:security" class="size-full" />
            <IconifyIcon v-else-if="row.meta?.icon" :icon="row.meta?.icon || 'carbon:circle-dash'" class="size-full" />
          </div>
          <span class="flex-auto">{{ $t(row.meta?.title) }}</span>
          <div class="items-center justify-end"></div>
        </div>
        <!-- <MenuBadge
          v-if="row.meta?.badgeType"
          class="menu-badge"
          :badge="row.meta.badge"
          :badge-type="row.meta.badgeType"
          :badge-variants="row.meta.badgeVariants"
        /> -->
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
