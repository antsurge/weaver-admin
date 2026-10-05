<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { SystemTenantApi } from '#/api/system/tenant';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import { createTenantApi, updateTenantApi } from '#/api/system/tenant';
import { $t } from '#/locales';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<SystemTenantApi.Tenant>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'code',
    label: $t('system.tenant.fields.code'),
    component: 'Input',
    rules: z
      .string()
      .min(1, $t('ui.formRules.required', [$t('system.tenant.fields.code')]))
      .max(
        64,
        $t('ui.formRules.maxLength', [$t('system.tenant.fields.code'), 64]),
      )
      .regex(/^[\w-]+$/, $t('system.tenant.formRules.codePattern')),
    componentProps: {
      placeholder: $t('system.tenant.formPlaceholder.code'),
      disabled: !!formData.value?.id,
    },
    // 编辑时编码不可修改
    dependencies: {
      triggerFields: [],
      show: () => !formData.value?.id,
    },
  },
  {
    fieldName: 'name',
    label: $t('system.tenant.fields.name'),
    component: 'Input',
    rules: z
      .string()
      .min(1, $t('ui.formRules.required', [$t('system.tenant.fields.name')]))
      .max(
        64,
        $t('ui.formRules.maxLength', [$t('system.tenant.fields.name'), 64]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.tenant.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'status',
    label: $t('system.tenant.fields.status'),
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
    fieldName: 'expireAt',
    label: $t('system.tenant.fields.expireAt'),
    component: 'DatePicker',
    componentProps: {
      showTime: true,
      style: { width: '100%' },
      placeholder: $t('system.tenant.formPlaceholder.expireAt'),
    },
  },
  {
    fieldName: 'maxUsers',
    label: $t('system.tenant.fields.maxUsers'),
    component: 'InputNumber',
    defaultValue: 0,
    componentProps: {
      min: 0,
      style: { width: '100%' },
      placeholder: $t('system.tenant.formPlaceholder.maxUsers'),
    },
  },
  {
    fieldName: 'maxRoles',
    label: $t('system.tenant.fields.maxRoles'),
    component: 'InputNumber',
    defaultValue: 0,
    componentProps: {
      min: 0,
      style: { width: '100%' },
      placeholder: $t('system.tenant.formPlaceholder.maxRoles'),
    },
  },
  {
    fieldName: 'contactName',
    label: $t('system.tenant.fields.contactName'),
    component: 'Input',
    rules: z
      .string()
      .max(
        64,
        $t('ui.formRules.maxLength', [
          $t('system.tenant.fields.contactName'),
          64,
        ]),
      )
      .optional(),
    componentProps: {
      placeholder: $t('system.tenant.formPlaceholder.contactName'),
    },
  },
  {
    fieldName: 'contactPhone',
    label: $t('system.tenant.fields.contactPhone'),
    component: 'Input',
    rules: z
      .string()
      .max(
        20,
        $t('ui.formRules.maxLength', [
          $t('system.tenant.fields.contactPhone'),
          20,
        ]),
      )
      .optional(),
    componentProps: {
      placeholder: $t('system.tenant.formPlaceholder.contactPhone'),
    },
  },
  {
    fieldName: 'remark',
    label: $t('system.tenant.fields.remark'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('system.tenant.formPlaceholder.remark'),
      autoSize: { minRows: 2, maxRows: 4 },
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
    const data = modalApi.getData<SystemTenantApi.Tenant>();
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
      SystemTenantApi.Tenant,
      'createdAt' | 'id' | 'updatedAt'
    >;

    await (formData.value?.id
      ? updateTenantApi(formData.value.id, data)
      : createTenantApi(data));

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('system.tenant.name')])
    : $t('ui.actionTitle.create', [$t('system.tenant.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[620px]" :title="getModalTitle">
    <Form :layout="isHorizontal ? 'horizontal' : 'vertical'" class="mx-4" />
  </Modal>
</template>
