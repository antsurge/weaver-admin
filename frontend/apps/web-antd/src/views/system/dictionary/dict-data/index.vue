<script setup lang="ts">
import { message, Modal } from 'ant-design-vue';
import { watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '#/locales';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { DictionaryDictDataApi } from '#/api/system/dictionary/dict-data';
import {
  deleteDictDataApi,
  getDictDataListApi,
  updateDictDataStatusApi,
} from '#/api/system/dictionary/dict-data';

import DictDataForm from '#/views/system/dictionary/dict-data/modules/form/index.vue';

import { useDictDataColumns } from './data';

const props = defineProps<{
  data: DictionaryDictDataApi.DictData[];
}>();

const [DictDataFormModal, dictDataFormApi] = useVbenModal({
  connectedComponent: DictDataForm,
  destroyOnClose: true,
});

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: useDictDataColumns(onActionClick, onStatusChange),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: false,
    },
    data: props.data,
    rowConfig: {
      keyField: 'id',
    },
  } as VxeTableGridOptions,
});

watch(
  () => props.data,
  (data) => {
    gridApi.setGridOptions({ data: data ?? [] });
  },
  { deep: true },
);

const reloadGrid = async () => {
  const dictTypeID = props.data?.[0]?.dictTypeID;
  if (!dictTypeID) return;

  const res = await getDictDataListApi({
    dictTypeID:dictTypeID,
  });
  const list =
    (Array.isArray(res) ? res : (res as any)?.items) ??
    (Array.isArray((res as any)?.list) ? (res as any).list : []);

  gridApi.setGridOptions({ data: list ?? [] });
};

const onDictDataFormSuccess = () => {
  void reloadGrid();
};

function onActionClick({
  code,
  row,
}: {
  code: string;
  row: DictionaryDictDataApi.DictData;
}) {
  switch (code) {
    case 'edit':
      dictDataFormApi.setData(row).open();
      break;
    case 'delete':
      void deleteDictDataApi([row.id]).then(async () => {
        message.success($t('ui.actionMessage.deleteSuccess', [row.label]));
        await reloadGrid();
      });
      break;
    default:
      break;
  }
}

function onStatusChange(
  newStatus: DictionaryDictDataApi.DictData['status'],
  row: DictionaryDictDataApi.DictData,
): Promise<boolean | undefined> {
  const statusText: Record<string, string> = {
    enabled: '启用',
    disabled: '禁用',
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: '切换状态',
      content: `你要将 ${row.label} 的状态切换为 【${
        statusText[newStatus]
      }】 吗？`,
      async onOk() {
        try {
          await updateDictDataStatusApi(row.id, newStatus);
          await reloadGrid();
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
</script>
<template>
  <DictDataFormModal @success="onDictDataFormSuccess" />
  <Grid></Grid>
</template>
