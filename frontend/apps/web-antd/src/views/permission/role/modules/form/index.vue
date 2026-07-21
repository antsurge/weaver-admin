<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm } from '#/adapter/form';
import {
  getRoleApi,
  createRoleApi,
  updateRoleApi,
} from '#/api/permission/role';
import type { PermissionRoleApi } from '#/api/permission/role';

import { $t } from '#/locales';

import { nameRule, codeRule } from './rules';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<PermissionRoleApi.Role>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('permission.role.fields.name'),
    component: 'Input',
    rules: nameRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('permission.role.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'code',
    label: $t('permission.role.fields.code'),
    component: 'Input',
    rules: codeRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('permission.role.fields.code'),
      ]),
    },
  },
  {
    fieldName: 'weight',
    label: $t('permission.role.fields.weight'),
    component: 'InputNumber',
    defaultValue: 0,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('permission.role.fields.weight'),
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
    label: $t('permission.role.fields.status'),
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
    label: $t('permission.role.fields.remark'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('permission.role.form_placeholder.remark'),
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

    const data = modalApi.getData<PermissionRoleApi.Role>();
    // 编辑
    if (data?.id) {
      modalApi.lock();
      try {
        const res = await getRoleApi(data.id);
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
  const data = await formApi.getValues<PermissionRoleApi.Role>();
  try {
    if (formData.value?.id) {
      await updateRoleApi(formData.value.id, data);
    } else {
      await createRoleApi(data);
    }
    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('permission.role.name')])
    : $t('ui.actionTitle.create', [$t('permission.role.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[720px]" :title="getModalTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'" />
  </Modal>
</template>
