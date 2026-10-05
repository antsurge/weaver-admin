<script lang="ts" setup>
import type { FormSchemaApi } from '#/api/lowcode/form-schema';

import { computed } from 'vue';

import { Table } from 'ant-design-vue';

import { $t } from '#/locales';

const props = defineProps<{
  fieldMap: Record<string, string>;
  list: FormSchemaApi.FormSubmission[];
  loading: boolean;
  page: number;
  pageSize: number;
  total: number;
}>();

const emit = defineEmits<{
  pageChange: [page: number];
  pageSizeChange: [size: number];
}>();

const columns = computed(() => {
  const keys = Object.keys(props.fieldMap).slice(0, 4);
  return [
    ...keys.map((key) => ({
      title: props.fieldMap[key] || key,
      dataIndex: ['submitData', key],
      ellipsis: true,
    })),
    {
      title: $t('lowcode.formBuilder.submission.submitTime'),
      dataIndex: 'createdAt',
      width: 180,
      ellipsis: true,
    },
  ];
});

function formatJson(record: FormSchemaApi.FormSubmission) {
  try {
    return JSON.stringify(record.submitData ?? {}, null, 2);
  } catch {
    return String(record.submitData ?? '');
  }
}
</script>

<template>
  <Table
    :data-source="list"
    :columns="columns"
    :loading="loading"
    :pagination="{
      current: page,
      pageSize,
      total,
      showSizeChanger: true,
      showTotal: (t: number) => $t('lowcode.formBuilder.submission.total', [t]),
    }"
    :row-key="(r) => r.id"
    size="small"
    @change="
      (pagination: any) => {
        emit('pageChange', pagination.current);
        if (pagination.pageSize !== pageSize) {
          emit('pageSizeChange', pagination.pageSize);
        }
      }
    "
  >
    <template #expandedRowRender="{ record }">
      <div class="px-4 py-2">
        <div class="mb-1 text-xs font-medium text-gray-500">
          {{ $t('lowcode.formBuilder.submission.submitData') }}
        </div>
        <pre
          class="max-h-[320px] overflow-auto rounded-md bg-gray-50 p-3 text-xs leading-relaxed dark:bg-gray-900"
          >{{ formatJson(record) }}</pre>
      </div>
    </template>
  </Table>
</template>
