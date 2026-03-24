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

const formData = ref<SystemDictApi.DictData>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'dictTypeId',
    label: $t('system.dict_data.fields.dictTypeId'),
    component: 'ApiSelect',
    componentProps: {
      api: async () => {
        const { getDictTypeListApi } = await import('#/api/system/dict');
        const res = await getDictTypeListApi();
        const list = res?.data || res;
        return (list || []).map((item: SystemDictApi.DictType) => ({
          label: item.name,
          value: item.id,
        }));
      },
      showSearch: true,
      filterOption(input: string, option: any) {
        return option.label.toLowerCase().includes(input.toLowerCase());
      },
      placeholder: $t('system.dict_data.form_placeholder.dictTypeId'),
    },
    rules: z.string({
      required_error: $t('ui.formRules.required', [
        $t('system.dict_data.fields.dictTypeId'),
      ]),
    }),
  },
  {
    fieldName: 'label',
    label: $t('system.dict_data.fields.label'),
    component: 'Input',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.minLength', [$t('system.dict_data.fields.label'), 1]),
      )
      .max(
        50,
        $t('ui.formRules.maxLength', [
          $t('system.dict_data.fields.label'),
          50,
        ]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.dict_data.fields.label'),
      ]),
    },
  },
  {
    fieldName: 'value',
    label: $t('system.dict_data.fields.value'),
    component: 'Input',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.minLength', [$t('system.dict_data.fields.value'), 1]),
      )
      .max(
        100,
        $t('ui.formRules.maxLength', [
          $t('system.dict_data.fields.value'),
          100,
        ]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.dict_data.fields.value'),
      ]),
    },
  },
  {
    fieldName: 'description',
    label: $t('system.dict_data.fields.description'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('system.dict_data.form_placeholder.description'),
      autoSize: { minRows: 2, maxRows: 4 },
    },
  },
  {
    fieldName: 'status',
    label: $t('system.dict_data.fields.status'),
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
    labelWidth: 90,
  },
  schema,
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange(isOpen) {
    if (!isOpen) return;
    const data = modalApi.getData<SystemDictApi.DictData & {
      dictTypeId?: string;
    }>();

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

    const { createDictDataApi, updateDictDataApi } = await import(
      '#/api/system/dict'
    );

    if (formData.value?.id) {
      await updateDictDataApi(formData.value.id, data);
    } else {
      await createDictDataApi(data);
    }

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('system.dict_data.name')])
    : $t('ui.actionTitle.create', [$t('system.dict_data.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[600px]" :title="getModalTitle">
    <Form :layout="isHorizontal ? 'horizontal' : 'vertical'" class="mx-4" />
  </Modal>
</template>

