<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm } from '#/adapter/form';
import {
  getDepartmentTreeApi,
  getDepartmentApi,
  createDepartmentApi,
  updateDepartmentApi,
} from '#/api/organization/department';

import type {
  OrganizationDepartmentApi,
} from '#/api/organization/department';


import { $t } from '#/locales';

import {
  nameRule,
  parentIDRule,
  codeRule,
} from "./rules"

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<OrganizationDepartmentApi.Department>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'parentID',
    label: $t('organization.department.fields.parentID'),
    component: 'ApiTreeSelect',
    rules: parentIDRule,
    componentProps: {
      api: getDepartmentTreeApi,
      allowClear: true,
      class: 'w-full',
      showSearch: true,
      treeDefaultExpandAll: true,
      labelField: 'label',
      valueField: 'id',
      childrenField: 'children',
      getPopupContainer,
      placeholder: $t('organization.department.form_placeholder.parentID'),
      afterFetch: (res: any[]) => {
        const list = (res as any)?.items || res;
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
    label: $t('organization.department.fields.name'),
    component: 'Input',
    rules: nameRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('organization.department.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'code',
    label: $t('organization.department.fields.code'),
    component: 'Input',
    rules: codeRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('organization.department.fields.code'),
      ]),
    },
  },
  {
    fieldName: 'weight',
    label: $t('organization.department.fields.weight'),
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
    label: $t('organization.department.fields.leaderName'),
    component: 'Input',
    componentProps: {
      placeholder: $t('organization.department.form_placeholder.leaderName'),
    },
  },
  {
    fieldName: 'leaderPhone',
    label: $t('organization.department.fields.leaderPhone'),
    component: 'Input',
    componentProps: {
      placeholder: $t('organization.department.form_placeholder.leaderPhone'),
    },
  },
  {
    fieldName: 'leaderEmail',
    label: $t('organization.department.fields.leaderEmail'),
    component: 'Input',
    componentProps: {
      placeholder: $t('organization.department.form_placeholder.leaderEmail'),
    },
  },
  {
    fieldName: 'status',
    label: $t('organization.department.fields.status'),
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
  onOpenChange: async (isOpen) => {
    if (!isOpen) return;

    const data = modalApi.getData<OrganizationDepartmentApi.Department>();

    if (data?.id) {
      modalApi.lock();
      try {
        const res = await getDepartmentApi(data.id);
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

  const data = await formApi.getValues<OrganizationDepartmentApi.Department>();

  try {
    if (formData.value?.id) {
      await updateDepartmentApi(formData.value.id, data);
    } else {
      await createDepartmentApi(data);
    }

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('organization.department.name')])
    : $t('ui.actionTitle.create', [$t('organization.department.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[720px]" :title="getModalTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'" />
  </Modal>
</template>
