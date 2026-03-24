<script setup lang="ts">
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { useDictDataColumns } from './data';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemDictApi } from "#/api/system/dict"

const props = defineProps<{
  data: SystemDictApi.DictData[]
}>()


const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: useDictDataColumns(),
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
</script>
<template>
  <Grid></Grid>
</template>
