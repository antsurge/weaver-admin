<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm} from '#/adapter/form';
import {
  createOrganization,
  getOrganizationTreeApi,
  updateOrganization,
} from '#/api/system/organization';

import type {
  SystemOrganizationApi,
} from '#/api/system/organization';


import { $t } from '#/locales';

import {
  nameRule,
  parentIdRule,
  codeRule,
} from "./rules"

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<SystemOrganizationApi.Organization>();





const schema: VbenFormSchema[] = [
  {
    fieldName: 'parentId',
    label: $t('system.organization.fields.parentId'),
    component: 'ApiTreeSelect',
    rules: parentIdRule,
    componentProps: {
      api: getOrganizationTreeApi,
      allowClear: true,
      class: 'w-full',
      showSearch: true,
      treeDefaultExpandAll: true,
      labelField: 'label',
      valueField: 'id',
      childrenField: 'children',
      getPopupContainer,
      placeholder: $t('system.organization.form_placeholder.parentId'),
      afterFetch: (res: any[]) => {
        const list = (res as any)?.data || res;
        const convert = (items: any[]): any[] =>
          items.map((item) => ({
            id: item.id,
            label: item.name,
            children: convert(item.children || []),
          }));
        return convert(list || []);
      },
    },
  },
  {
    fieldName: 'name',
    label: $t('system.organization.fields.name'),
    component: 'Input',
    rules: nameRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.organization.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'code',
    label: $t('system.organization.fields.code'),
    component: 'Input',
    rules: codeRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.organization.fields.code'),
      ]),
    },
  },
  {
    fieldName: 'weight',
    label: $t('system.organization.fields.weight'),
    component: 'InputNumber',
    componentProps: {
      min: 0,
      step: 1,
      precision: 0,
      stringMode: false,
      style: { width: '100%' },
    },
  },
  {
    fieldName: 'leaderName',
    label: $t('system.organization.fields.leaderName'),
    component: 'Input',
    componentProps: {
      placeholder: $t('system.organization.form_placeholder.leaderName'),
    },
  },
  {
    fieldName: 'leaderPhone',
    label: $t('system.organization.fields.leaderPhone'),
    component: 'Input',
    componentProps: {
      placeholder: $t('system.organization.form_placeholder.leaderPhone'),
    },
  },
  {
    fieldName: 'leaderEmail',
    label: $t('system.organization.fields.leaderEmail'),
    component: 'Input',
    componentProps: {
      placeholder: $t('system.organization.form_placeholder.leaderEmail'),
    },
  },
  {
    fieldName: 'status',
    label: $t('system.organization.fields.status'),
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
    formItemClass: 'col-span-2 md:col-span-2',
    labelWidth: 90,
  },
  schema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2 gap-x-4',
});

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange(isOpen) {
    if (!isOpen) return;

    const data = modalApi.getData<SystemOrganizationApi.Organization>();

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

  const data = await formApi.getValues<SystemOrganizationApi.Organization>();

  try {
    if (formData.value?.id) {
      await updateOrganization(formData.value.id, data);
    } else {
      await createOrganization(data);
    }

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('system.organization.name')])
    : $t('ui.actionTitle.create', [$t('system.organization.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[720px]" :title="getModalTitle">
    <Form
      class="mx-4"
      :layout="isHorizontal ? 'horizontal' : 'vertical'"
    />
  </Modal>
</template>

