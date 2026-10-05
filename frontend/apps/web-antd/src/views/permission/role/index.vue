<script lang="ts" setup>
import type { Recordable } from '@vben/types';

import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { PermissionRoleApi } from '#/api/permission/role';

import { nextTick, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { IconifyIcon, Plus } from '@vben/icons';
import { $t } from '@vben/locales';

import { Button, message, Modal } from 'ant-design-vue';

import { PermissionAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteRoleApi,
  getRoleListApi,
  updateRoleStatusApi,
} from '#/api/permission/role';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import { useColumns, useFormOptions } from './data';
import Form from './modules/form/index.vue';

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
  const checkboxRecords = gridApi.grid?.getCheckboxRecords?.() ?? [];
  const checkboxReserveRecords =
    gridApi.grid?.getCheckboxReserveRecords?.() ?? [];
  selectedRows.value = [...checkboxRecords, ...checkboxReserveRecords];
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
  // 超级管理员、系统内置角色不允许编辑
  if (row.isSuperAdmin) {
    message.warning($t('permission.role.actions.superAdminNotAllowed'));
    return;
  }
  if (row.isSystem) {
    message.warning($t('permission.role.actions.systemNotAllowed'));
    return;
  }
  formModelApi.setData(row).open();
}

function onCreate() {
  formModelApi.setData({}).open();
}

function onStatusChange(
  newStatus: PermissionRoleApi.Role['status'],
  row: PermissionRoleApi.Role,
): Promise<boolean | undefined> {
  // 超级管理员、系统内置角色不允许修改状态
  if (row.isSuperAdmin) {
    message.warning($t('permission.role.actions.superAdminNotAllowed'));
    return Promise.resolve(false);
  }
  if (row.isSystem) {
    message.warning($t('permission.role.actions.systemNotAllowed'));
    return Promise.resolve(false);
  }

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

function onDelete(row: PermissionRoleApi.Role) {
  // 超级管理员、系统内置角色不允许删除
  if (row.isSuperAdmin) {
    message.warning($t('permission.role.actions.superAdminNotAllowed'));
    return;
  }
  if (row.isSystem) {
    message.warning($t('permission.role.actions.systemNotAllowed'));
    return;
  }
  deleteRoleApi([row.id])
    .then(() => {
      onRefresh();
    })
    .catch(() => {});
}

function onBatchDelete() {
  const rows = selectedRows.value;
  if (!rows?.length) return;
  // 过滤掉超级管理员与系统内置角色，不允许删除
  const deletableRows = rows.filter((r) => !r.isSystem && !r.isSuperAdmin);
  if (deletableRows.length === 0) {
    message.warning($t('permission.role.actions.systemNotAllowed'));
    return;
  }
  const ids = deletableRows.map((r) => r.id);
  const names = deletableRows.map((r) => r.name || r.id).join('、');

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
          <Button
            v-access:code="[PermissionAuthCode.Role.BatchDelete]"
            type="primary"
            danger
            :disabled="selectedRows.length === 0"
            @click="onBatchDelete"
          >
            <IconifyIcon icon="ant-design:delete-outlined" class="size-5" />
            {{ $t('ui.actionTitle.delete') }}
          </Button>
          <Button
            v-access:code="[PermissionAuthCode.Role.Create]"
            type="primary"
            @click="onCreate"
          >
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>
        </div>
      </template>
    </Grid>
  </Page>
</template>
