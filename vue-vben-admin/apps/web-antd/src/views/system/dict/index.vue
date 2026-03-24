<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemDictApi } from '#/api/system/dict';

import { nextTick, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';
import { $t } from '@vben/locales';

import { Button, message, Modal } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteDictDataApi,
  deleteDictTypeApi,
  getDictTypeListApi,
  updateDictDataStatusApi,
  updateDictTypeStatusApi,
} from '#/api/system/dict';

import { useDictDataColumns, useDictTypeColumns } from './data';
import DictTypeForm from './modules/type-form/index.vue';
import DictDataForm from './modules/data-form/index.vue';

const [DictTypeFormModal, dictTypeFormApi] = useVbenModal({
  connectedComponent: DictTypeForm,
  destroyOnClose: true,
});

const [DictDataFormModal, dictDataFormApi] = useVbenModal({
  connectedComponent: DictDataForm,
  destroyOnClose: true,
});

const selectedTypes = ref<SystemDictApi.DictType[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: useDictTypeColumns(onTypeActionClick as any, onTypeStatusChange),
    height: 'auto',
    // keepSource: true,

    // ✅ 核心：控制是否显示展开按钮
    expandConfig: {
      // visibleMethod: ({ row }: any) => {
      //   return Array.isArray(row.dictData) && row.dictData.length > 0;
      // },
    },

    pagerConfig: {
      enabled: false,
    },
    proxyConfig: {
      ajax: {
        query: async () => {
          const res = await getDictTypeListApi();
          return res?.data ?? res ?? [];
        },
      },
    },

    rowConfig: {
      keyField: 'id',
    },

    checkboxConfig: { range: false },

    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
    },
  } as VxeTableGridOptions,

  gridEvents: {
    checkboxChange() {
      nextTick(() => {
        const records =
          (gridApi.grid as any)?.getCheckboxRecords?.() ?? [];
        selectedTypes.value = records;
      });
    },
  },
});

function onRefresh() {
  gridApi.query();
}

function onCreateType() {
  dictTypeFormApi.setData({}).open();
}

function onCreateData(dictTypeId?: string) {
  dictDataFormApi
    .setData(dictTypeId ? { dictTypeId } : {})
    .open();
}

function onTypeActionClick({
  code,
  row,
}: {
  code: string;
  row: SystemDictApi.DictType;
}) {
  switch (code) {
    case 'appendDictData':
      onCreateData(row.id);
      break;
    case 'edit':
      dictTypeFormApi.setData(row).open();
      break;
    case 'delete':
      onDeleteTypes([row]);
      break;
  }
}

function onDictDataActionClick({
  code,
  row,
}: {
  code: string;
  row: SystemDictApi.DictData;
}) {
  switch (code) {
    case 'edit':
      dictDataFormApi.setData(row).open();
      break;
    case 'delete':
      onDeleteData([row]);
      break;
  }
}

async function onTypeStatusChange(
  newStatus: SystemDictApi.DictType['status'],
  row: SystemDictApi.DictType,
) {
  const statusText: Record<string, string> = {
    enabled: '启用',
    disabled: '禁用',
  };

  return new Promise((resolve) => {
    Modal.confirm({
      title: '切换状态',
      content: `你要将 ${row.name} 的状态切换为 【${statusText[newStatus]}】 吗？`,
      async onOk() {
        try {
          await updateDictTypeStatusApi(row.id, newStatus);
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

async function onDictDataStatusChange(
  newStatus: SystemDictApi.DictData['status'],
  row: SystemDictApi.DictData,
) {
  const statusText: Record<string, string> = {
    enabled: '启用',
    disabled: '禁用',
  };

  return new Promise((resolve) => {
    Modal.confirm({
      title: '切换状态',
      content: `你要将 ${row.label} 的状态切换为 【${statusText[newStatus]}】 吗？`,
      async onOk() {
        try {
          await updateDictDataStatusApi(row.id, newStatus);
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

function onDeleteTypes(rows?: SystemDictApi.DictType[]) {
  const list = rows?.length ? rows : selectedTypes.value;
  if (!list.length) return;

  const ids = list.map((item) => item.id);
  const names = list.map((item) => item.name).join('、');

  Modal.confirm({
    title: $t('ui.actionMessage.confirmDelete'),
    content: $t('ui.actionMessage.deleteConfirm', [names]),
    okType: 'danger',
    async onOk() {
      const hideLoading = message.loading({
        content: $t('ui.actionMessage.deleting', [names]),
        duration: 0,
      });
      try {
        await deleteDictTypeApi(ids);
        message.success($t('ui.actionMessage.deleteSuccess', [names]));
        selectedTypes.value = [];
        onRefresh();
      } catch {
        hideLoading();
      }
    },
  });
}

function onDeleteData(rows: SystemDictApi.DictData[]) {
  if (!rows.length) return;

  const ids = rows.map((item) => item.id);
  const labels = rows.map((item) => item.label).join('、');

  Modal.confirm({
    title: $t('ui.actionMessage.confirmDelete'),
    content: $t('ui.actionMessage.deleteConfirm', [labels]),
    okType: 'danger',
    async onOk() {
      const hideLoading = message.loading({
        content: $t('ui.actionMessage.deleting', [labels]),
        duration: 0,
      });
      try {
        await deleteDictDataApi(ids);
        message.success($t('ui.actionMessage.deleteSuccess', [labels]));
        onRefresh();
      } catch {
        hideLoading();
      }
    },
  });
}

import type { VxeGridPropTypes } from 'vxe-table'

interface RowVO {
  id: number
  name: string
  role: string
  sex: string
  age: number
  address: string
}

const tableColumn = ref<VxeGridPropTypes.Columns>([
  { type: 'seq', width: 70 },
  { field: 'name', title: 'Name' },
  { field: 'sex', title: 'Sex' },
  { field: 'age', title: 'Age' }
])

const tableData = ref<RowVO[]>([
  { id: 10001, name: 'Test1', role: 'Develop', sex: 'Man', age: 28, address: 'test abc' },
  { id: 10002, name: 'Test2', role: 'Test', sex: 'Women', age: 22, address: 'Guangzhou' },
  { id: 10003, name: 'Test3', role: 'PM', sex: 'Man', age: 32, address: 'Shanghai' },
  { id: 10004, name: 'Test4', role: 'Designer', sex: 'Women', age: 24, address: 'Shanghai' }
])
</script>

<template>
  <Page auto-content-height>
    <DictTypeFormModal @success="onRefresh" />
    <DictDataFormModal @success="onRefresh" />

    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            type="primary"
            danger
            @click="onDeleteTypes()"
            :disabled="!selectedTypes.length"
          >
            {{ $t('ui.actionTitle.delete') }}
          </Button>

          <Button type="primary" @click="onCreateType">
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>
        </div>
      </template>

      <template #expand_dictdata="{ row }">
        <div class="p-4 bg-[var(--vben-color-bg-elevated)]" style="height:100px">
          <vxe-grid :columns="tableColumn" :data="tableData" min-height="200"
          show-header="true" ></vxe-grid>
        </div>
      </template>
    </Grid>
  </Page>
</template>
