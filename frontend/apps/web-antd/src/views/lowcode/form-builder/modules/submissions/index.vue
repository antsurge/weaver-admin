<script lang="ts" setup>
import type { FormSchemaApi } from '#/api/lowcode/form-schema';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Empty } from 'ant-design-vue';

import { getFormSubmissionListApi } from '#/api/lowcode/form-schema';
import { $t } from '#/locales';

import FormSubmissionTable from './table.vue';

const form = ref<FormSchemaApi.FormSchema>();
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const loading = ref(false);
const list = ref<FormSchemaApi.FormSubmission[]>([]);

const [Modal, modalApi] = useVbenModal({
  footer: false,
  onOpenChange(isOpen) {
    if (!isOpen) return;
    const data = modalApi.getData<FormSchemaApi.FormSchema>();
    if (!data) return;
    form.value = data;
    load();
  },
});

const fieldMap = computed(() => {
  const schemaJson = form.value?.schemaJson as
    undefined | { fields?: { key: string; label: string }[] };
  const map: Record<string, string> = {};
  (schemaJson?.fields ?? []).forEach((f) => {
    map[f.key] = f.label;
  });
  return map;
});

async function load() {
  if (!form.value?.code) return;
  loading.value = true;
  try {
    const res = await getFormSubmissionListApi(form.value.code, {
      currentPage: page.value,
      pageSize: pageSize.value,
    });
    list.value = res.items ?? [];
    total.value = Number(res.total ?? 0);
  } finally {
    loading.value = false;
  }
}

function onPageChange(p: number) {
  page.value = p;
  load();
}

function onPageSizeChange(size: number) {
  pageSize.value = size;
  page.value = 1;
  load();
}
</script>

<template>
  <Modal
    class="w-full max-w-[860px]"
    :title="$t('lowcode.formBuilder.submission.title', [form?.name ?? ''])"
  >
    <FormSubmissionTable
      :list="list"
      :field-map="fieldMap"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="pageSize"
      @page-change="onPageChange"
      @page-size-change="onPageSizeChange"
    />
    <div v-if="!loading && list.length === 0" class="py-10">
      <Empty :description="$t('lowcode.formBuilder.submission.empty')" />
    </div>
  </Modal>
</template>
