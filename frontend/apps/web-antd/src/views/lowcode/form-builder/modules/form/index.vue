<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { FormSchemaApi } from '#/api/lowcode/form-schema';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import {
  createFormSchemaApi,
  updateFormSchemaApi,
} from '#/api/lowcode/form-schema';
import { $t } from '#/locales';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<FormSchemaApi.FormSchema>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('lowcode.formBuilder.fields.name'),
    component: 'Input',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.required', [$t('lowcode.formBuilder.fields.name')]),
      )
      .max(
        50,
        $t('ui.formRules.maxLength', [
          $t('lowcode.formBuilder.fields.name'),
          50,
        ]),
      ),
    componentProps: {
      placeholder: $t('lowcode.formBuilder.fields.name'),
    },
  },
  {
    fieldName: 'code',
    label: $t('lowcode.formBuilder.fields.code'),
    component: 'Input',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.required', [$t('lowcode.formBuilder.fields.code')]),
      )
      .max(
        100,
        $t('ui.formRules.maxLength', [
          $t('lowcode.formBuilder.fields.code'),
          100,
        ]),
      )
      .regex(
        /^[a-z][\w-]*$/i,
        $t('ui.formRules.pattern', [$t('lowcode.formBuilder.fields.code')]),
      ),
    componentProps: {
      placeholder: $t('lowcode.formBuilder.fields.code'),
    },
  },
  {
    fieldName: 'description',
    label: $t('lowcode.formBuilder.fields.description'),
    component: 'Input',
    componentProps: {
      placeholder: $t('lowcode.formBuilder.fields.description'),
    },
  },
  {
    fieldName: 'status',
    label: $t('lowcode.formBuilder.fields.status'),
    component: 'Select',
    defaultValue: 'enabled',
    componentProps: {
      options: [
        { label: $t('ui.status.enabled'), value: 'enabled' },
        { label: $t('ui.status.disabled'), value: 'disabled' },
      ],
    },
  },
];

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

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
    const data = modalApi.getData<FormSchemaApi.FormSchema>();
    if (data) {
      formData.value = data;
      formApi.setValues(data);
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
    const data = (await formApi.getValues()) as Omit<
      FormSchemaApi.FormSchema,
      'createdAt' | 'id' | 'updatedAt'
    >;

    await (formData.value?.id
      ? updateFormSchemaApi(formData.value.id, data)
      : createFormSchemaApi(data));

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('lowcode.formBuilder.name')])
    : $t('ui.actionTitle.create', [$t('lowcode.formBuilder.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[600px]" :title="getModalTitle">
    <Form :layout="isHorizontal ? 'horizontal' : 'vertical'" class="mx-4" />
  </Modal>
</template>
