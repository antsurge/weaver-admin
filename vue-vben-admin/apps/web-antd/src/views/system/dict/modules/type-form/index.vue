<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { SystemDictApi } from '#/api/system/dict';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import { $t } from '#/locales';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<SystemDictApi.DictType>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('system.dict_type.fields.name'),
    component: 'Input',
    rules: z
      .string()
      .min(
        2,
        $t('ui.formRules.minLength', [$t('system.dict_type.fields.name'), 2]),
      )
      .max(
        50,
        $t('ui.formRules.maxLength', [$t('system.dict_type.fields.name'), 50]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.dict_type.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'code',
    label: $t('system.dict_type.fields.code'),
    component: 'Input',
    rules: z
      .string()
      .min(
        2,
        $t('ui.formRules.minLength', [$t('system.dict_type.fields.code'), 2]),
      )
      .max(
        50,
        $t('ui.formRules.maxLength', [$t('system.dict_type.fields.code'), 50]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.dict_type.fields.code'),
      ]),
    },
  },
  {
    fieldName: 'description',
    label: $t('system.dict_type.fields.description'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('system.dict_type.form_placeholder.description'),
      autoSize: { minRows: 2, maxRows: 4 },
    },
  },
  {
    fieldName: 'status',
    label: $t('system.dict_type.fields.status'),
    component: 'Switch',
    defaultValue: 'enabled',
    componentProps: {
      class: 'w-auto',
      checkedChildren: $t('common.enabled'),
      checkedValue: 'enabled',
      unCheckedChildren: $t('common.disabled'),
      unCheckedValue: 'disabled',
    },
  },
];

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    colon: true,
    labelWidth: 80,
  },
  schema,
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange(isOpen) {
    if (!isOpen) return;
    const data = modalApi.getData<SystemDictApi.DictType>();
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
    const data = await formApi.getValues();

    const { createDictTypeApi, updateDictTypeApi } = await import(
      '#/api/system/dict'
    );

    if (formData.value?.id) {
      await updateDictTypeApi(formData.value.id, data);
    } else {
      await createDictTypeApi(data);
    }

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('system.dict_type.name')])
    : $t('ui.actionTitle.create', [$t('system.dict_type.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[600px]" :title="getModalTitle">
    <Form :layout="isHorizontal ? 'horizontal' : 'vertical'" class="mx-4" />
  </Modal>
</template>

