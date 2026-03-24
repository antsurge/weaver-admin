<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm } from '#/adapter/form';
import {
  getPositionApi,
  createPositionApi,
  updatePositionApi,
} from '#/api/system/position';
import type { SystemPositionApi } from '#/api/system/position';

import { $t } from '#/locales';

import { nameRule, codeRule } from './rules';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<SystemPositionApi.Position>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('system.position.fields.name'),
    component: 'Input',
    rules: nameRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.position.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'code',
    label: $t('system.position.fields.code'),
    component: 'Input',
    rules: codeRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.position.fields.code'),
      ]),
    },
  },
  {
    fieldName: 'weight',
    label: $t('system.position.fields.weight'),
    component: 'InputNumber',
    defaultValue: 0,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.position.fields.weight'),
      ]),
      min: 0,
      step: 1,
      precision: 0,
      stringMode: false,
      style: { width: '100%' },
    },
  },
  {
    fieldName: 'status',
    label: $t('system.position.fields.status'),
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
  {
    fieldName: 'remark',
    label: $t('system.position.fields.remark'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('system.position.form_placeholder.remark'),
    },
  },
];

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    colon: true,
    formItemClass: 'col-span-2 md:col-span-2',
    labelWidth: 90,
  },
  schema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2 gap-x-4',
});

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange:async (isOpen) => {
    if (!isOpen) return;

    const data = modalApi.getData<SystemPositionApi.Position>();
    // 编辑
    if (data?.id) {
      modalApi.lock(); // 加个loading体验更好
      try {
        const res = await getPositionApi(data.id);
        formData.value = res;
        formApi.setValues(res);
      } finally {
        modalApi.unlock();
      }
    } else {
      // 👉 新增
      formData.value = undefined;
      formApi.resetForm();
    }
  },
});

async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) return;
  modalApi.lock();
  const data = await formApi.getValues<SystemPositionApi.Position>();
  try {
    if (formData.value?.id) {
      await updatePositionApi(formData.value.id, data);
    } else {
      await createPositionApi(data);
    }
    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('system.position.name')])
    : $t('ui.actionTitle.create', [$t('system.position.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[720px]" :title="getModalTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'" />
  </Modal>
</template>
