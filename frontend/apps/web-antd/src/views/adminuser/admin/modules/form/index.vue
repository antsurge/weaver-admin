<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm } from '#/adapter/form';
import {
  getAdminApi,
  createAdminApi,
  updateAdminApi,
} from '#/api/adminuser/admin';
import type { AdminuserAdminApi } from '#/api/adminuser/admin';

import { $t } from '#/locales';

import { nameRule, codeRule } from './rules';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<AdminuserAdminApi.Admin>();

  const schema: VbenFormSchema[] = [
  {
    fieldName: 'realName',
    label: '真实姓名',
    component: 'Input',
    // rules: [{ required: true, message: '请输入真实姓名' }],
    componentProps: {
      placeholder: '请输入真实姓名',
    },
  },
  {
    fieldName: 'username',
    label: '用户名',
    component: 'Input',
    // rules: [{ required: true, message: '请输入用户名' }],
    componentProps: {
      placeholder: '请输入用户名',
    },
  },
  {
    fieldName: 'email',
    label: '邮箱',
    component: 'Input',
    // rules: [
    //   { type: 'email', message: '请输入正确的邮箱格式' },
    // ],
    componentProps: {
      placeholder: '请输入邮箱（可选）',
    },
  },
  {
    fieldName: 'phone',
    label: '手机号',
    component: 'Input',
    // rules: [
    //   {
    //     pattern: /^1\d{10}$/,
    //     message: '请输入正确的手机号',
    //   },
    // ],
    componentProps: {
      placeholder: '请输入手机号（可选）',
    },
  },
  {
    fieldName: 'avatar',
    label: '头像',
    component: 'Input', // 👉 如果你有上传组件可以换成 Upload
    componentProps: {
      placeholder: '请输入头像地址（可选）',
    },
  },
  {
    fieldName: 'status',
    label: '状态',
    component: 'Switch',
    defaultValue: 'enabled',
    componentProps: {
      class: 'w-auto',
      checkedChildren: '启用',
      checkedValue: 'enabled',
      unCheckedChildren: '禁用',
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
  onOpenChange:async (isOpen) => {
    if (!isOpen) return;

    const data = modalApi.getData<AdminuserAdminApi.Admin>();
    // 编辑
    if (data?.id) {
      modalApi.lock();
      try {
        const res = await getAdminApi(data.id);
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
  const data = await formApi.getValues<AdminuserAdminApi.Admin>();
  try {
    if (formData.value?.id) {
      await updateAdminApi(formData.value.id, data);
    } else {
      await createAdminApi(data);
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
