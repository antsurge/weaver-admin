<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemConfigApi } from '#/api/system/config';

import { nextTick, onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal, Tag } from 'ant-design-vue';

import { SystemAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteConfigApi,
  getConfigListApi,
  updateConfigStatusApi,
} from '#/api/system/config';
import { $t } from '#/locales';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import { useConfigColumns, useFormOptions } from './data';
import ConfigForm from './modules/form/index.vue';

const [ConfigFormModal, configFormApi] = useVbenModal({
  connectedComponent: ConfigForm,
  destroyOnClose: true,
});

const selectedRows = ref<SystemConfigApi.Config[]>([]);

// 分组选项（供搜索筛选使用，从全量参数中提取去重）
const groupOptions = ref<{ label: string; value: string }[]>([]);

async function loadGroupOptions() {
  try {
    const res = await getConfigListApi({ currentPage: 1, pageSize: 0 });
    const groups = new Set<string>();
    (res?.items ?? []).forEach((item) => {
      if (item.group) groups.add(item.group);
    });
    groupOptions.value = [...groups].map((g) => ({ label: g, value: g }));
  } catch {
    groupOptions.value = [];
  }
}

onMounted(loadGroupOptions);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(groupOptions),
  gridOptions: {
    columns: useConfigColumns(onActionClick, onStatusChange),
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
          const res = await getConfigListApi({
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
  // 同步刷新分组选项（新建/修改分组后搜索下拉能立即选中）
  loadGroupOptions();
}

function onCreate() {
  configFormApi.setData({}).open();
}

function onActionClick({
  code,
  row,
}: {
  code: string;
  row: SystemConfigApi.Config;
}) {
  switch (code) {
    case 'delete': {
      onDelete([row]);
      break;
    }
    case 'edit': {
      configFormApi.setData(row).open();
      break;
    }
  }
}

function onStatusChange(
  newStatus: SystemConfigApi.Config['status'],
  row: SystemConfigApi.Config,
): Promise<boolean | undefined> {
  const statusText: Record<string, string> = {
    enabled: $t('common.enabled'),
    disabled: $t('common.disabled'),
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: $t('common.actions.confirm'),
      content: `${$t('system.config.message.statusConfirm')}【${statusText[newStatus]}】吗？`,
      async onOk() {
        try {
          await updateConfigStatusApi(row.id, newStatus);
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

function onDelete(rows: SystemConfigApi.Config[]) {
  Modal.confirm({
    title: $t('system.config.confirm.deleteTitle'),
    content: $t('system.config.confirm.deleteContent', [rows.length]),
    async onOk() {
      const ids = rows.map((item) => item.id);
      await deleteConfigApi(ids);
      message.success($t('system.config.message.deleteSuccess'));
      onRefresh();
    },
  });
}

function onBatchDelete() {
  if (selectedRows.value.length === 0) return;
  onDelete(selectedRows.value);
}
</script>

<template>
  <Page auto-content-height>
    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            v-access:code="[SystemAuthCode.Config.Create]"
            type="primary"
            @click="onCreate"
          >
            <Plus class="mr-1" />
            {{ $t('common.actions.create') }}
          </Button>
          <Button
            v-access:code="[SystemAuthCode.Config.BatchDelete]"
            danger
            :disabled="selectedRows.length === 0"
            @click="onBatchDelete"
          >
            {{ $t('system.config.operation.batchDelete') }}
          </Button>
        </div>
      </template>
      <template #config_group="{ row }">
        <Tag v-if="row.group" color="cyan">{{ row.group }}</Tag>
        <span v-else class="text-gray-400">
          {{ $t('system.config.config_group_options.none') }}
        </span>
      </template>
      <template #config_type="{ row }">
        <Tag :color="row.configType === 'Y' ? 'gold' : 'blue'">
          {{
            row.configType === 'Y'
              ? $t('system.config.config_type_options.Y')
              : $t('system.config.config_type_options.N')
          }}
        </Tag>
      </template>
    </Grid>
    <ConfigFormModal @success="onRefresh" />
  </Page>
</template>
