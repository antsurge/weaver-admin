<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { SystemJobApi } from '#/api/system/job';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import { createJobApi, updateJobApi } from '#/api/system/job';
import { $t } from '#/locales';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<SystemJobApi.Job>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('system.job.fields.name'),
    component: 'Input',
    rules: z
      .string()
      .min(2, $t('ui.formRules.minLength', [$t('system.job.fields.name'), 2]))
      .max(
        50,
        $t('ui.formRules.maxLength', [$t('system.job.fields.name'), 50]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [$t('system.job.fields.name')]),
    },
  },
  {
    fieldName: 'jobGroup',
    label: $t('system.job.fields.jobGroup'),
    component: 'Input',
    defaultValue: 'DEFAULT',
    componentProps: {
      placeholder: $t('system.job.form_placeholder.jobGroup'),
    },
  },
  {
    fieldName: 'invokeTarget',
    label: $t('system.job.fields.invokeTarget'),
    component: 'Input',
    rules: z
      .string()
      .min(
        2,
        $t('ui.formRules.minLength', [$t('system.job.fields.invokeTarget'), 2]),
      )
      .max(
        200,
        $t('ui.formRules.maxLength', [
          $t('system.job.fields.invokeTarget'),
          200,
        ]),
      ),
    componentProps: {
      placeholder: $t('system.job.form_placeholder.invokeTarget'),
    },
  },
  {
    fieldName: 'cronExpression',
    label: $t('system.job.fields.cronExpression'),
    component: 'Input',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.required', [$t('system.job.fields.cronExpression')]),
      )
      .max(
        50,
        $t('ui.formRules.maxLength', [
          $t('system.job.fields.cronExpression'),
          50,
        ]),
      ),
    componentProps: {
      placeholder: $t('system.job.form_placeholder.cronExpression'),
    },
  },
  {
    fieldName: 'misfirePolicy',
    label: $t('system.job.fields.misfirePolicy'),
    component: 'Select',
    defaultValue: 'immediately',
    componentProps: {
      options: [
        {
          label: $t('system.job.misfire_options.immediately'),
          value: 'immediately',
        },
        {
          label: $t('system.job.misfire_options.once'),
          value: 'once',
        },
        {
          label: $t('system.job.misfire_options.ignore'),
          value: 'ignore',
        },
      ],
    },
  },
  {
    fieldName: 'concurrent',
    label: $t('system.job.fields.concurrent'),
    component: 'Switch',
    defaultValue: false,
    componentProps: {
      class: 'w-auto',
      checkedChildren: $t('common.yes'),
      unCheckedChildren: $t('common.no'),
    },
  },
  {
    fieldName: 'status',
    label: $t('system.job.fields.status'),
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
    label: $t('system.job.fields.remark'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('system.job.form_placeholder.remark'),
      autoSize: { minRows: 2, maxRows: 4 },
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
    const data = modalApi.getData<SystemJobApi.Job>();
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
      SystemJobApi.Job,
      'createdAt' | 'id' | 'updatedAt'
    >;

    // 任务类型目前仅支持 http，表单未暴露该字段，提交前补默认值
    // （proto 层 CEL 校验 jobType 必须为 'http'，缺失会报 job.jobType.invalid）
    data.jobType = data.jobType || 'http';

    await (formData.value?.id
      ? updateJobApi(formData.value.id, data)
      : createJobApi(data));

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('system.job.name')])
    : $t('ui.actionTitle.create', [$t('system.job.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[600px]" :title="getModalTitle">
    <Form :layout="isHorizontal ? 'horizontal' : 'vertical'" class="mx-4" />
  </Modal>
</template>
