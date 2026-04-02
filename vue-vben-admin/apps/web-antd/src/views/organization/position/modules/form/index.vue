<script lang="ts" setup>
// ==================== types ====================
import type { VbenFormSchema } from '#/adapter/form';
import type { OrganizationPositionApi } from '#/api/organization/position';

// ==================== vue ====================
import { computed, ref } from 'vue';

// ==================== vben ====================
import { useVbenForm, z } from '#/adapter/form';
import { useVbenModal } from '@vben/common-ui';
import { $t } from '#/locales';

// ==================== third-party ====================
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

// ==================== api ====================
import {
  getPositionApi,
  createPositionApi,
  updatePositionApi,
  isPositionNameExistsApi,
  isPositionCodeExistsApi,
} from '#/api/organization/position';


const emit = defineEmits<{
  success: [];
}>();

const formData = ref<OrganizationPositionApi.Position>();
const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('organization.position.fields.name'),
    component: 'Input',
    rules: z
      .string()
      .min(2, $t('ui.formRules.minLength', [$t('organization.position.fields.name'), 2]))
      .max(30, $t('ui.formRules.maxLength', [$t('organization.position.fields.name'), 30]))
      .refine(
        async (value: string) => {
          if (!value) {
            return false
          }
          let res = await isPositionNameExistsApi(value, formData.value?.id)
          return !res?.exists;
        },
        (value) => ({
          message: $t('ui.formRules.alreadyExists', [
            $t('organization.position.fields.name'),
            value,
          ]),
        }),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('organization.position.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'code',
    label: $t('organization.position.fields.code'),
    component: 'Input',
    rules: z
      .string()
      .min(2, $t('ui.formRules.minLength', [$t('organization.position.fields.code'), 2]))
      .max(30, $t('ui.formRules.maxLength', [$t('organization.position.fields.code'), 30]))
      .refine(
        async (value: string) => {
          if (!value) {
            return false
          }
          let res = await isPositionCodeExistsApi(value, formData.value?.id)
          return !res?.exists;
        },
        (value) => ({
          message: $t('ui.formRules.alreadyExists', [
            $t('organization.position.fields.code'),
            value,
          ]),
        }),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('organization.position.fields.code'),
      ]),
    },
  },
  {
    fieldName: 'weight',
    label: $t('organization.position.fields.weight'),
    component: 'InputNumber',
    defaultValue: 0,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('organization.position.fields.weight'),
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
    label: $t('organization.position.fields.status'),
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
    label: $t('organization.position.fields.remark'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('organization.position.form_placeholder.remark'),
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
  schema: schema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2 gap-x-4',
});

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange: async (isOpen) => {
    if (!isOpen) return;

    const data = modalApi.getData<OrganizationPositionApi.Position>();
    // 编辑
    if (data?.id) {
      modalApi.lock();
      try {
        const res = await getPositionApi(data.id);
        formData.value = res;
        formApi.setValues(res);
      } finally {
        modalApi.unlock();
      }
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
  const data = await formApi.getValues<OrganizationPositionApi.Position>();
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
    ? $t('ui.actionTitle.edit', [$t('organization.position.name')])
    : $t('ui.actionTitle.create', [$t('organization.position.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[720px]" :title="getModalTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'" />
  </Modal>
</template>
