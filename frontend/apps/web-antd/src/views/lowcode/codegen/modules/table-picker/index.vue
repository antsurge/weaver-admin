<script lang="ts" setup>
import type { CodegenApi } from '#/api/lowcode/codegen';
import type { AllResult } from '#/types/pagination';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Form, FormItem, message, Select } from 'ant-design-vue';

import { getDbTableListApi } from '#/api/lowcode/codegen';
import { $t } from '#/locales';

const emit = defineEmits<{
  select: [table: CodegenApi.DbTable];
}>();

const loading = ref(false);
const selectedTableName = ref<string>();
const tableOptions = ref<
  { label: string; table: CodegenApi.DbTable; value: string }[]
>([]);
let searchTimer: number | undefined;

async function loadTables(keyword?: string) {
  loading.value = true;
  let res: AllResult<CodegenApi.DbTable> | undefined;
  try {
    res = await getDbTableListApi(keyword?.trim() || undefined);
  } catch (error) {
    message.error(
      (error as Error)?.message || $t('lowcode.codegen.message.loadTableFail'),
    );
  } finally {
    loading.value = false;
  }
  const items = res?.items ?? [];
  tableOptions.value = items.map((t) => ({
    label: t.tableComment ? `${t.tableName}（${t.tableComment}）` : t.tableName,
    value: t.tableName,
    table: t,
  }));
}

// 服务端搜索（防抖）
function onSearch(keyword: string) {
  window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(() => loadTables(keyword), 300);
}

function onConfirm() {
  const selected = tableOptions.value.find(
    (o) => o.value === selectedTableName.value,
  )?.table;
  if (!selected) {
    message.warning($t('lowcode.codegen.message.selectTableFirst'));
    return;
  }
  modalApi.close();
  emit('select', selected);
}

const [Modal, modalApi] = useVbenModal({
  onConfirm,
  onOpenChange(isOpen) {
    if (isOpen && tableOptions.value.length === 0) loadTables();
  },
});
</script>

<template>
  <Modal
    class="w-full max-w-[480px]"
    :title="$t('lowcode.codegen.selectTable')"
  >
    <div class="px-4 py-5">
      <Form layout="vertical">
        <FormItem :label="$t('lowcode.codegen.fields.tableName')">
          <Select
            v-model:value="selectedTableName"
            :options="tableOptions"
            :loading="loading"
            :filter-option="false"
            allow-clear
            show-search
            :placeholder="$t('lowcode.codegen.selectTablePlaceholder')"
            @search="onSearch"
          />
          <div
            v-if="!loading && tableOptions.length === 0"
            class="mt-1 text-xs text-gray-400"
          >
            {{ $t('lowcode.codegen.noTable') }}
          </div>
        </FormItem>
      </Form>
    </div>
  </Modal>
</template>
