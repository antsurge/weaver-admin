<script lang="ts" setup>
import type { Recordable } from '@vben/types';

import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { AdminuserAdminApi } from '#/api/adminuser/admin';

import { h, nextTick, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { IconifyIcon, Plus } from '@vben/icons';
import { $t } from '@vben/locales';

import { Button, Image, Input, message, Modal, Tag } from 'ant-design-vue';

import { AdminAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteAdminApi,
  exportAdminApi,
  getAdminListApi,
  resetPasswordApi,
  updateAdminStatusApi,
} from '#/api/adminuser/admin';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';
import {
  downloadFile,
  getFileNameFromDisposition,
  handleBlobResponseError,
} from '#/utils/download';

import Form from '../form/index.vue';
import { useColumns, useFormOptions } from './data';

// 左侧部门树选中的部门（用于列表过滤 + 新增时默认归属）
const props = defineProps<{
  departmentId?: string;
  departmentName?: string;
}>();

const [FormModel, formModelApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

const selectedRows = ref<AdminuserAdminApi.Admin[]>([]);

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
          const res = await getAdminListApi({
            ...page,
            ...formValues,
            // 左侧选中部门时按部门过滤
            ...(props.departmentId ? { departmentId: props.departmentId } : {}),
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
}: OnActionClickParams<AdminuserAdminApi.Admin>) {
  switch (code) {
    case 'delete': {
      onDelete(row);
      break;
    }
    case 'edit': {
      onEdit(row);
      break;
    }
    case 'resetPassword': {
      onResetPassword(row);
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

// 左侧选中部门变化时重新查询列表
watch(
  () => props.departmentId,
  () => {
    onRefresh();
  },
);

function onEdit(row: AdminuserAdminApi.Admin) {
  formModelApi.setData(row).open();
}

function onCreate() {
  // 左侧选中部门时，新增用户默认选中该部门（可在表单树中改选）
  formModelApi.setData(
    props.departmentId ? { departmentId: props.departmentId } : {},
  );
  formModelApi.open();
}

function onStatusChange(
  newStatus: AdminuserAdminApi.Admin['status'],
  row: AdminuserAdminApi.Admin,
): Promise<boolean | undefined> {
  const statusText: Recordable<string> = {
    disabled: $t('common.disabled'),
    enabled: $t('common.enabled'),
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: $t('adminuser.admin.actions.switchStatus'),
      content: $t('adminuser.admin.actions.switchStatusConfirm', [
        row.username,
        statusText[newStatus],
      ]),
      okText: $t('adminuser.admin.actions.confirm'),
      cancelText: $t('adminuser.admin.actions.cancel'),
      async onOk() {
        try {
          await updateAdminStatusApi(row.id, newStatus);
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

function onDelete(row: AdminuserAdminApi.Admin) {
  if (row.isSuperAdmin) {
    message.warning($t('adminuser.admin.actions.superAdminNotAllowed'));
    return;
  }
  Modal.confirm({
    title: $t('adminuser.admin.actions.delete'),
    content: $t('adminuser.admin.actions.deleteConfirm', [
      row.username || row.realName,
    ]),
    okText: $t('adminuser.admin.actions.confirm'),
    cancelText: $t('adminuser.admin.actions.cancel'),
    okType: 'danger',
    async onOk() {
      await deleteAdminApi([row.id]);
      onRefresh();
    },
  });
}

// 重置密码：弹出密码输入框，提交后调用后端接口
function onResetPassword(row: AdminuserAdminApi.Admin) {
  const password = ref('');
  Modal.confirm({
    title: $t('adminuser.admin.actions.resetPassword'),
    content: () =>
      h(Input.Password, {
        value: password.value,
        placeholder: $t('adminuser.admin.actions.resetPasswordPlaceholder'),
        'onUpdate:value': (val: string) => {
          password.value = val;
        },
      }),
    okText: $t('adminuser.admin.actions.confirm'),
    cancelText: $t('adminuser.admin.actions.cancel'),
    async onOk() {
      const value = password.value?.trim();
      if (!value) {
        message.warning($t('adminuser.admin.actions.resetPasswordRequired'));
        throw new Error('empty password');
      }
      await resetPasswordApi(row.id, value);
    },
  });
}

// 导出当前查询条件下的用户列表
async function onExport() {
  try {
    const formValues = await gridApi.formApi?.getValues?.();
    const res = await exportAdminApi(formValues);
    const filename = getFileNameFromDisposition(
      res.headers?.['content-disposition'],
    );
    downloadFile(
      res.data,
      filename || `用户列表_${new Date().toISOString().slice(0, 10)}.csv`,
    );
  } catch (error: any) {
    handleBlobResponseError(error);
  }
}

function onBatchDelete() {
  const rows = selectedRows.value;
  if (!rows?.length) return;
  // 超级管理员账号不允许删除，批量操作时直接排除
  const deletableRows = rows.filter((r) => !r.isSuperAdmin);
  if (deletableRows.length === 0) {
    message.warning($t('adminuser.admin.actions.superAdminNotAllowed'));
    return;
  }
  const ids = deletableRows.map((r) => r.id);

  Modal.confirm({
    title: $t('adminuser.admin.actions.delete'),
    content: $t('adminuser.admin.actions.deleteBatchConfirm', [ids.length]),
    okText: $t('adminuser.admin.actions.confirm'),
    cancelText: $t('adminuser.admin.actions.cancel'),
    okType: 'danger',
    onOk() {
      const hideLoading = message.loading({
        content: $t('adminuser.admin.actions.deleting'),
        duration: 0,
        key: 'batch_delete_msg',
      });
      return deleteAdminApi(ids)
        .then(() => {
          message.success({
            content: $t('adminuser.admin.actions.deleteSuccess'),
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
  <FormModel @success="onRefresh" />
  <Grid>
    <template #avatar="{ row }">
      <Image
        v-if="row.avatarUrl"
        :src="row.avatarUrl"
        :width="36"
        :height="36"
        :preview="true"
        class="cursor-pointer rounded-full"
      />
    </template>
    <template #roleNames="{ row }">
      <template v-if="row.isSuperAdmin">
        <Tag color="gold">{{ $t('adminuser.admin.superAdmin') }}</Tag>
      </template>
      <template v-else-if="row.roleNames?.length">
        <Tag v-for="name in row.roleNames" :key="name" color="blue">
          {{ name }}
        </Tag>
      </template>
      <span v-else>-</span>
    </template>
    <template #dataScope="{ row }">
      <template v-if="row.dataPermissionNames?.length">
        <Tag v-for="name in row.dataPermissionNames" :key="name" color="cyan">
          {{ name }}
        </Tag>
      </template>
      <span v-else class="text-gray-400">{{
        $t('adminuser.admin.followRole')
      }}</span>
    </template>
    <template #toolbar-tools>
      <div class="flex gap-2">
        <Button
          v-access:code="[AdminAuthCode.Admin.BatchDelete]"
          type="primary"
          class="inline-flex items-center"
          danger
          :disabled="selectedRows.length === 0"
          @click="onBatchDelete"
        >
          <IconifyIcon icon="ant-design:delete-outlined" class="size-5" />
          {{ $t('ui.actionTitle.delete') }}
        </Button>
        <Button
          v-access:code="[AdminAuthCode.Admin.Export]"
          @click="onExport"
          class="inline-flex items-center"
        >
          <IconifyIcon icon="ant-design:download-outlined" class="size-5" />
          {{ $t('adminuser.admin.actions.export') }}
        </Button>
        <Button
          v-access:code="[AdminAuthCode.Admin.Create]"
          type="primary"
          @click="onCreate"
          class="inline-flex items-center"
        >
          <Plus class="size-5" />
          {{ $t('ui.actionTitle.create') }}
        </Button>
      </div>
    </template>
  </Grid>
</template>
