<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui'
import { markRaw, ref} from 'vue'
import { AuthenticationLogin, z } from '@vben/common-ui'
import { $t } from '@vben/locales'
import CaptchaInput from './captcha-input.vue'
import { useAuthStore } from '#/store'

defineOptions({ name: 'Login' })

const authStore = useAuthStore()

const captchaId = ref("")

const refreshTrigger = ref(0);

// 表单 schema 使用稳定常量，避免整表单被重新创建
const formSchema: VbenFormSchema[] = [
  {
    component: 'VbenInput',
    fieldName: 'username',
    label: $t('authentication.username'),
    componentProps: { placeholder: $t('authentication.usernameTip') },
    rules: z.string().min(1, { message: $t('authentication.usernameTip') })
  },
  {
    component: 'VbenInputPassword',
    fieldName: 'password',
    label: $t('authentication.password'),
    componentProps: { placeholder: $t('authentication.password') },
    rules: z.string().min(1, { message: $t('authentication.passwordTip') })
  },
  {
    component: markRaw(CaptchaInput),
    fieldName: 'captcha',
    label: $t('authentication.captcha'),
    componentProps: {
      placeholder: $t('authentication.captchaTip'),
      refreshTrigger: refreshTrigger,
      "onUpdate:captchaId":(id:string) => {
        captchaId.value = id
      }
    },
    rules: z.string().min(1, { message: $t('authentication.captchaTip') })
  }
]

// 登录处理
const handleLogin = async (values: any) => {
  try {
    const payload = { 
      ...values, 
      captchaId:captchaId.value
    }
    await authStore.authLogin(payload)
  } catch (error) {
    // 刷新验证码
    refreshTrigger.value = Date.now();
  }
}

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
