<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { NotificationApi } from '#/api/system/notification';

import { computed, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';
import { message, Select } from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import { getAdminListApi } from '#/api/adminuser/admin';
import { getRoleListApi } from '#/api/permission/role';
import { createNotificationApi } from '#/api/system/notification';
import { $t } from '#/locales';

import {
  getNotificationLevelOptions,
  getNotificationTargetOptions,
  getNotificationTypeOptions,
} from '../../data';

const emit = defineEmits<{
  success: [];
}>();

// ───────────── 基础字段（交给 VbenForm） ─────────────

const schema: VbenFormSchema[] = [
  {
    fieldName: 'title',
    label: $t('system.notification.fields.title'),
    component: 'Input',
    rules: z
      .string()
      .min(
        1,
        $t('ui.formRules.required', [$t('system.notification.fields.title')]),
      )
      .max(
        128,
        $t('ui.formRules.maxLength', [
          $t('system.notification.fields.title'),
          128,
        ]),
      ),
    componentProps: {
      placeholder: $t('system.notification.form_placeholder.title'),
    },
  },
  {
    fieldName: 'content',
    label: $t('system.notification.fields.content'),
    component: 'Textarea',
    rules: z
      .string()
      .max(
        1024,
        $t('ui.formRules.maxLength', [
          $t('system.notification.fields.content'),
          1024,
        ]),
      ),
    componentProps: {
      placeholder: $t('system.notification.form_placeholder.content'),
      autoSize: { minRows: 3, maxRows: 6 },
    },
  },
  {
    fieldName: 'type',
    label: $t('system.notification.fields.type'),
    component: 'Select',
    defaultValue: 'system',
    componentProps: {
      class: 'w-full',
      options: getNotificationTypeOptions(),
    },
  },
  {
    fieldName: 'level',
    label: $t('system.notification.fields.level'),
    component: 'Select',
    defaultValue: 'info',
    componentProps: {
      class: 'w-full',
      options: getNotificationLevelOptions(),
    },
  },
  {
    fieldName: 'link',
    label: $t('system.notification.fields.link'),
    component: 'Input',
    componentProps: {
      placeholder: $t('system.notification.form_placeholder.link'),
    },
  },
];

// ───────────── 投放范围（联动，单独用原生 Select 处理） ─────────────

type TargetType = 'all' | 'role' | 'user';

const targetType = ref<TargetType>('all');
const targetIds = ref<string[]>([]);
const targetOptions = ref<{ label: string; value: string }[]>([]);
const targetLoading = ref(false);

const needTarget = computed(() => targetType.value !== 'all');

async function loadTargetOptions() {
  if (!needTarget.value) {
    targetOptions.value = [];
    return;
  }
  targetLoading.value = true;
  try {
    if (targetType.value === 'role') {
      const res = await getRoleListApi({ currentPage: 1, pageSize: 1000 });
      targetOptions.value = (res.items ?? []).map((item) => ({
        label: item.name,
        value: item.id,
      }));
    } else {
      const res = await getAdminListApi({ currentPage: 1, pageSize: 1000 });
      targetOptions.value = (res.items ?? []).map((item) => ({
        label: item.realName
          ? `${item.realName}(${item.username})`
          : item.username,
        value: item.id,
      }));
    }
  } finally {
    targetLoading.value = false;
  }
}

// 切换投放范围时清空已选目标并重新拉取选项
watch(targetType, () => {
  targetIds.value = [];
  void loadTargetOptions();
});

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
    formApi.resetForm();
    targetType.value = 'all';
    targetIds.value = [];
    targetOptions.value = [];
  },
});

async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) return;

  if (needTarget.value && targetIds.value.length === 0) {
    // 选择了指定用户/角色却没有选目标，直接拦下
    return;
  }

  modalApi.lock();
  try {
    const values = await formApi.getValues<NotificationApi.CreateParams>();
    await createNotificationApi({
      ...values,
      targetType: targetType.value,
      targetIds: needTarget.value ? targetIds.value : [],
    });
    message.success($t('system.notification.message.publishSuccess'));
    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  $t('ui.actionTitle.create', [$t('system.notification.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[640px]" :title="getModalTitle">
    <div class="mx-4">
      <Form :layout="isHorizontal ? 'horizontal' : 'vertical'" />

      <!-- 投放范围 -->
      <div class="border-t pt-4">
        <div class="mb-3 flex items-center gap-1 text-base font-medium">
          {{ $t('system.notification.fields.targetType') }}
        </div>
        <Select
          v-model:value="targetType"
          class="w-full"
          :options="getNotificationTargetOptions()"
        />

        <template v-if="needTarget">
          <div class="mb-3 mt-4 flex items-center gap-1 text-base font-medium">
            {{ $t('system.notification.fields.targetIds') }}
          </div>
          <Select
            v-model:value="targetIds"
            mode="multiple"
            class="w-full"
            allow-clear
            show-search
            option-filter-prop="label"
            :loading="targetLoading"
            :options="targetOptions"
            :placeholder="$t('system.notification.form_placeholder.targetIds')"
          />
        </template>
      </div>
    </div>
  </Modal>
</template>
