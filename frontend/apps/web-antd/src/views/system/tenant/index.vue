<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemTenantApi } from '#/api/system/tenant';

import { nextTick, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';
import { $t } from '@vben/locales';

import { Button, message, Modal } from 'ant-design-vue';

import { SystemAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteTenantApi,
  getTenantListApi,
  updateTenantStatusApi,
} from '#/api/system/tenant';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import { useFormOptions, useTenantColumns } from './data';
import TenantForm from './modules/form/index.vue';

const [TenantFormModal, tenantFormApi] = useVbenModal({
  connectedComponent: TenantForm,
  destroyOnClose: true,
});

const selectedTenants = ref<SystemTenantApi.Tenant[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useTenantColumns(onTenantActionClick, onTenantStatusChange),
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
          const res = await getTenantListApi({
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
        selectedTenants.value = records;
      });
    },
  },
});

function onRefresh() {
  gridApi.query();
}

function onCreateTenant() {
  tenantFormApi.setData({}).open();
}

function onTenantActionClick({
  code,
  row,
}: {
  code: string;
  row: SystemTenantApi.Tenant;
}) {
  switch (code) {
    case 'delete': {
      onDeleteTenants([row]);
      break;
    }
    case 'edit': {
      tenantFormApi.setData(row).open();
      break;
    }
  }
}

function onTenantStatusChange(
  newStatus: SystemTenantApi.Tenant['status'],
  row: SystemTenantApi.Tenant,
): Promise<boolean | undefined> {
  const statusText: Record<string, string> = {
    enabled: $t('common.enabled'),
    disabled: $t('common.disabled'),
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: $t('system.tenant.confirm.switchStatusTitle'),
      content: $t('system.tenant.confirm.switchStatusContent', [
        row.name,
        statusText[newStatus],
      ]),
      async onOk() {
        try {
          await updateTenantStatusApi(row.id, newStatus);
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

async function onDeleteTenants(rows?: SystemTenantApi.Tenant[]) {
  const list = rows?.length ? rows : selectedTenants.value;
  if (list.length === 0) return;

  const ids = list.map((item) => item.id);
  const names = list.map((item) => item.name).join('、');

  await deleteTenantApi(ids);
  message.success($t('ui.actionMessage.deleteSuccess', [names]));
  selectedTenants.value = [];
  onRefresh();
}
</script>

<template>
  <Page auto-content-height>
    <TenantFormModal @success="onRefresh" />

    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            v-access:code="[SystemAuthCode.Tenant.Create]"
            type="primary"
            @click="onCreateTenant"
          >
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>
        </div>
      </template>
    </Grid>
  </Page>
</template>
