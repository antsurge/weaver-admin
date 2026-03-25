<script setup lang="ts">
import { watch } from 'vue';

import type {
  OnActionClickFn,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { SystemDictApi } from '#/api/system/dict';

import { useVbenVxeGrid } from '#/adapter/vxe-table';

import { useDictDataColumns } from './data';

const props = defineProps<{
  data: SystemDictApi.DictData[];
  onActionClick: OnActionClickFn<SystemDictApi.DictData>;
  onStatusChange: (
    newStatus: SystemDictApi.DictData['status'],
    row: SystemDictApi.DictData,
  ) => PromiseLike<boolean | undefined>;
}>();

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: useDictDataColumns(props.onActionClick, props.onStatusChange),
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
</script>
<template>
  <Grid></Grid>
</template>
