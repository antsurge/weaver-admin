<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed, markRaw, ref } from 'vue';

import { AuthenticationLogin, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { useAuthStore } from '#/store';

import CaptchaInput from './captcha-input.vue';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();

const captchaId = ref('');

const refreshTrigger = ref(0);

// 租户模式：multi=显示租户编码输入框（由构建期环境变量控制，默认 single）
const isMultiTenant = import.meta.env.VITE_TENANT_MODE === 'multi';

// 表单 schema 使用稳定常量，避免整表单被重新创建
const formSchema = computed<VbenFormSchema[]>(() => {
  const schemas: VbenFormSchema[] = [
    {
      component: 'VbenInput',
      fieldName: 'username',
      label: $t('authentication.username'),
      componentProps: { placeholder: $t('authentication.usernameTip') },
      rules: z.string().min(1, { message: $t('authentication.usernameTip') }),
    },
    {
      component: 'VbenInputPassword',
      fieldName: 'password',
      label: $t('authentication.password'),
      componentProps: { placeholder: $t('authentication.password') },
      rules: z.string().min(1, { message: $t('authentication.passwordTip') }),
    },
  ];
  if (isMultiTenant) {
    schemas.splice(0, 0, {
      component: 'VbenInput',
      fieldName: 'tenantCode',
      label: $t('authentication.tenantCode'),
      componentProps: { placeholder: $t('authentication.tenantCodeTip') },
      rules: z.string().min(1, { message: $t('authentication.tenantCodeTip') }),
    });
  }
  schemas.push({
    component: markRaw(CaptchaInput),
    fieldName: 'captcha',
    label: $t('authentication.captcha'),
    componentProps: {
      placeholder: $t('authentication.captchaTip'),
      refreshTrigger,
      'onUpdate:captchaId': (id: string) => {
        captchaId.value = id;
      },
    },
    rules: z.string().min(1, { message: $t('authentication.captchaTip') }),
  });
  return schemas;
});

// 登录处理
const handleLogin = async (values: any) => {
  try {
    const payload = {
      ...values,
      captchaId: captchaId.value,
    };
    await authStore.authLogin(payload);
  } catch {
    // 刷新验证码
    refreshTrigger.value = Date.now();
  }
};
</script>

<template>
  <AuthenticationLogin
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    :show-forget-password="false"
    :show-qrcode-login="false"
    :show-register="false"
    :show-third-party-login="false"
    :show-code-login="false"
    @submit="handleLogin"
  />
</template>
