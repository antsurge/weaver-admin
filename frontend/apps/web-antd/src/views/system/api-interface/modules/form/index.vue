<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { SystemApiInterfaceApi } from '#/api/system/api-interface';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { useVbenForm, z } from '#/adapter/form';
import {
  createApiInterfaceApi,
  getApiInterfaceOptionsApi,
  updateApiInterfaceApi,
} from '#/api/system/api-interface';
import { $t } from '#/locales';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<SystemApiInterfaceApi.ApiInterface>();

// HTTP 方法选项
const METHOD_OPTIONS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(
  (item) => ({
    label: item,
    value: item,
  }),
);

// 服务名/标签下拉选项（ApiSelect 的 api 属性异步加载）
async function loadServiceOptions() {
  try {
    const res = await getApiInterfaceOptionsApi();
    const services = res?.services ?? [];
    return services.map((item) => ({ label: item, value: item }));
  } catch {
    return [];
  }
}

async function loadTagOptions() {
  try {
    const res = await getApiInterfaceOptionsApi();
    const tags = res?.tags ?? [];
    return tags.map((item) => ({ label: item, value: item }));
  } catch {
    return [];
  }
}

const schema: VbenFormSchema[] = [
  {
    fieldName: 'service',
    label: $t('system.api_interface.fields.service'),
    component: 'ApiSelect',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.required', [
          $t('system.api_interface.fields.service'),
        ]),
      ),
    componentProps: {
      placeholder: $t('system.api_interface.form_placeholder.service'),
      class: 'w-full',
      showSearch: true,
      allowClear: true,
      api: loadServiceOptions,
    },
  },
  {
    fieldName: 'tag',
    label: $t('system.api_interface.fields.tag'),
    component: 'ApiSelect',
    componentProps: {
      placeholder: $t('system.api_interface.form_placeholder.tag'),
      class: 'w-full',
      showSearch: true,
      allowClear: true,
      api: loadTagOptions,
    },
  },
  {
    fieldName: 'method',
    label: $t('system.api_interface.fields.method'),
    component: 'Select',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.required', [$t('system.api_interface.fields.method')]),
      ),
    componentProps: {
      placeholder: $t('system.api_interface.form_placeholder.method'),
      class: 'w-full',
      options: METHOD_OPTIONS,
    },
  },
  {
    fieldName: 'path',
    label: $t('system.api_interface.fields.path'),
    component: 'Input',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.required', [$t('system.api_interface.fields.path')]),
      ),
    componentProps: {
      placeholder: $t('system.api_interface.form_placeholder.path'),
    },
  },
  {
    fieldName: 'summary',
    label: $t('system.api_interface.fields.summary'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('system.api_interface.form_placeholder.summary'),
      autoSize: { minRows: 2, maxRows: 4 },
    },
  },
];

const [Form, formApi] = useVbenForm({
  commonConfig: {
    colon: true,
    labelWidth: 100,
  },
  schema,
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange(isOpen) {
    if (!isOpen) return;
    const data = modalApi.getData<SystemApiInterfaceApi.ApiInterface>();
    if (data) {
      formData.value = data;
      // 只回填可编辑字段
      formApi.setValues({
        service: data.service,
        tag: data.tag,
        method: data.method,
        path: data.path,
        summary: data.summary,
      });
    } else {
      formData.value = undefined;
      formApi.resetForm();
    }
  },
});

async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) return;

  modalApi.lock();
  try {
    const data =
      (await formApi.getValues()) as SystemApiInterfaceApi.ApiInterfaceFormData;

    await (formData.value?.id
      ? updateApiInterfaceApi(formData.value.id, data)
      : createApiInterfaceApi(data));

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('system.api_interface.name')])
    : $t('ui.actionTitle.create', [$t('system.api_interface.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[600px]" :title="getModalTitle">
    <Form class="mx-4" />
  </Modal>
</template>
