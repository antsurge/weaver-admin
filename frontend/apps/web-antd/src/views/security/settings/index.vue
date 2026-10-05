<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { SecurityApi } from '#/api/security/security';

import { computed, onMounted, ref } from 'vue';

import { useAccess } from '@vben/access';
import { Page } from '@vben/common-ui';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';
import { Button, Card, message, Spin } from 'ant-design-vue';

import { SecurityAuthCode } from '#/access/security';
import { useVbenForm, z } from '#/adapter/form';
import {
  getSecurityPolicyApi,
  updateSecurityPolicyApi,
} from '#/api/security/security';
import { $t } from '#/locales';

const loading = ref(false);
const saving = ref(false);

const { hasAccessByCodes } = useAccess();
/** 是否有编辑安全策略的权限 */
const canEdit = computed(() =>
  hasAccessByCodes([SecurityAuthCode.Policy.Edit]),
);

const schema: VbenFormSchema[] = [
  {
    fieldName: 'loginFailEnabled',
    label: $t('security.security_policy.fields.loginFailEnabled'),
    component: 'Switch',
    defaultValue: false,
    componentProps: {
      class: 'w-auto',
      checkedChildren: $t('common.enabled'),
      unCheckedChildren: $t('common.disabled'),
    },
    help: $t('security.security_policy.tips.lock'),
  },
  {
    fieldName: 'loginFailMax',
    label: $t('security.security_policy.fields.loginFailMax'),
    component: 'InputNumber',
    defaultValue: 5,
    rules: z.number().min(1).max(100),
    componentProps: {
      class: 'w-40',
      min: 1,
      max: 100,
    },
  },
  {
    fieldName: 'loginFailWindowMinutes',
    label: $t('security.security_policy.fields.loginFailWindowMinutes'),
    component: 'InputNumber',
    defaultValue: 15,
    rules: z.number().min(1).max(1440),
    componentProps: {
      class: 'w-40',
      min: 1,
      max: 1440,
    },
  },
  {
    fieldName: 'loginLockMinutes',
    label: $t('security.security_policy.fields.loginLockMinutes'),
    component: 'InputNumber',
    defaultValue: 15,
    rules: z.number().min(1).max(1440),
    componentProps: {
      class: 'w-40',
      min: 1,
      max: 1440,
    },
  },
  {
    fieldName: 'ipMode',
    label: $t('security.security_policy.fields.ipMode'),
    component: 'RadioGroup',
    defaultValue: 'off',
    componentProps: {
      options: [
        {
          label: $t('security.security_policy.ip_mode_options.off'),
          value: 'off',
        },
        {
          label: $t('security.security_policy.ip_mode_options.whitelist'),
          value: 'whitelist',
        },
        {
          label: $t('security.security_policy.ip_mode_options.blacklist'),
          value: 'blacklist',
        },
      ],
    },
    help: $t('security.security_policy.tips.ip'),
  },
  {
    fieldName: 'ipList',
    label: $t('security.security_policy.fields.ipList'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('security.security_policy.placeholder.ipList'),
      autoSize: { minRows: 3, maxRows: 8 },
      rows: 4,
    },
  },
  {
    fieldName: 'accessTokenTtlSeconds',
    label: $t('security.security_policy.fields.accessTokenTtlSeconds'),
    component: 'InputNumber',
    defaultValue: 7200,
    rules: z.number().min(60).max(2_592_000),
    componentProps: {
      class: 'w-48',
      min: 60,
      max: 2_592_000,
    },
    help: $t('security.security_policy.tips.token'),
  },
  {
    fieldName: 'refreshTokenTtlSeconds',
    label: $t('security.security_policy.fields.refreshTokenTtlSeconds'),
    component: 'InputNumber',
    defaultValue: 604_800,
    rules: z.number().min(300).max(31_536_000),
    componentProps: {
      class: 'w-48',
      min: 300,
      max: 31_536_000,
    },
  },
  {
    fieldName: 'remark',
    label: $t('security.security_policy.fields.remark'),
    component: 'Textarea',
    componentProps: {
      autoSize: { minRows: 2, maxRows: 4 },
    },
  },
];

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

/** 表单公共配置：无编辑权限时整体禁用 */
const commonConfig = computed(() => ({
  colon: true,
  labelWidth: 200,
  disabled: !canEdit.value,
}));

const [Form, formApi] = useVbenForm({
  schema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-1',
});

/** 把 textarea 的多行文本解析为 IP 列表 */
function parseIpList(text: string): string[] {
  return (text || '')
    .split(/[\n,]/)
    .map((v) => v.trim())
    .filter(Boolean);
}

async function loadPolicy() {
  loading.value = true;
  try {
    const policy = await getSecurityPolicyApi();
    await formApi.setValues({
      ...policy,
      ipList: (policy.ipList || []).join('\n'),
    });
  } catch (error) {
    console.error(error);
    message.error($t('security.security_policy.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

async function onSubmit() {
  if (!canEdit.value) return;
  const { valid } = await formApi.validate();
  if (!valid) return;

  saving.value = true;
  try {
    const values = await formApi.getValues();
    const payload: SecurityApi.SecurityPolicy = {
      loginFailEnabled: Boolean(values.loginFailEnabled),
      loginFailMax: Number(values.loginFailMax),
      loginFailWindowMinutes: Number(values.loginFailWindowMinutes),
      loginLockMinutes: Number(values.loginLockMinutes),
      ipMode: values.ipMode,
      ipList: parseIpList(values.ipList),
      accessTokenTtlSeconds: Number(values.accessTokenTtlSeconds),
      refreshTokenTtlSeconds: Number(values.refreshTokenTtlSeconds),
      remark: values.remark,
    };
    await updateSecurityPolicyApi(payload);
  } finally {
    saving.value = false;
  }
}

onMounted(loadPolicy);
</script>

<template>
  <Page auto-content-height>
    <Card :title="$t('security.security_policy.title')" :bordered="false">
      <Spin :spinning="loading">
        <div class="max-w-[820px]">
          <Form
            :layout="isHorizontal ? 'horizontal' : 'vertical'"
            :common-config="commonConfig"
          />
          <div class="mt-6 flex gap-3">
            <Button
              type="primary"
              :loading="saving"
              :disabled="!canEdit"
              @click="onSubmit"
            >
              {{ $t('common.actions.save') }}
            </Button>
            <Button :disabled="!canEdit" @click="loadPolicy">
              {{ $t('common.actions.reset') }}
            </Button>
          </div>
        </div>
      </Spin>
    </Card>
  </Page>
</template>
