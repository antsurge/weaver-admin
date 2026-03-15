<template>
    <div>
        <AuthToolbar />
        <div class="flex justify-center items-center h-screen">
            <a-form :model="form" :rules="rules" layout="vertical" class="login-form" @finish="onFinish">
                <!-- 用户名 -->
                <a-form-item name="username" :rules="[{ required: true, message: $t('login.username_required') }]">
                    <a-input v-model:value="form.username" :placeholder="$t('login.username')">
                        <template #prefix>
                            <user-outlined />
                        </template>
                    </a-input>
                </a-form-item>

                <!-- 密码 -->
                <a-form-item name="password" :rules="[{ required: true, message: $t('login.password_required') }]">
                    <a-input-password v-model:value="form.password" :placeholder="$t('login.password')">
                        <template #prefix>
                            <lock-outlined />
                        </template>
                    </a-input-password>
                </a-form-item>

                <!-- 图片验证码 -->
                <a-form-item name="captcha" :rules="[{ required: true, message: $t('login.captcha_required') }]">
                    <div class="captcha-wrapper">
                        <a-input v-model:value="form.captcha" :placeholder="$t('login.captcha')" class="captcha-input">
                            <template #prefix>
                                <safety-certificate-outlined />
                            </template>
                        </a-input>
                        <img :src="captcha.imageBase64" alt="验证码" class="captcha-img" @click="reloadCaptcha" />
                    </div>
                </a-form-item>

                <!-- 登录按钮 -->
                <a-form-item>
                    <a-button type="primary" html-type="submit" block size="large" :loading="loading">
                        {{ $t('login.submit') }}
                    </a-button>
                </a-form-item>
            </a-form>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { message } from "ant-design-vue";
import { useRouter } from 'vue-router';
// 引入图标
import { UserOutlined, LockOutlined, SafetyCertificateOutlined } from '@ant-design/icons-vue';
import { $t } from "@/locales"
// 引入国际化
import AuthToolbar from "@/layouts/auth/toolbar.vue"
import {
    login,
    getCaptcha,
} from "@/api/auth"

const loading = ref(false);
const router = useRouter();

// 表单数据
interface LoginForm {
    username: string;
    password: string;
    captcha: string;
}

const form = reactive<LoginForm>({
    username: "",
    password: "",
    captcha: "",
});

// 表单校验规则
const rules = reactive({
    username: [{ required: true, message: $t('login.username_required') }],
    password: [{ required: true, message: $t('login.password_required') }],
    // captcha: [{ required: true, message: $t('login.captcha_required') }],
});

// 验证码 URL
const captcha = ref({
    captchaId: "",
    imageBase64: ""
})

onMounted(async () => {
    reloadCaptcha()
})

// 刷新验证码
const reloadCaptcha = async () => {
    let res = await getCaptcha()
    captcha.value = res
}

async function onFinish() {
    loading.value = true; // 开启 loading

    const data = {
        ...form,
        captchaId: captcha.value.captchaId,
    };

    try {
        const res = await login(data);
        console.log("登录成功", res);
        message.success($t('login.success'));
        // 登录成功后跳转 Layout
        router.push('/dashboard'); // 默认进入 Layout 的首页
    } catch (err: any) {
        // 刷新验证码
        reloadCaptcha();
        // 提示错误信息
        message.error(err?.message || $t('login.failed'));
    } finally {
        loading.value = false; // 关闭 loading
    }
}
</script>

<style scoped lang="scss">
.login-form {
    max-width: 400px;
    margin: 0 auto;

    // 移除 label 后，表单项间距可能需要微调，Antd 默认会有 margin-bottom
    :deep(.ant-form-item) {
        margin-bottom: 24px;
    }
}

.captcha-wrapper {
    display: flex;
    align-items: center;
    width: 100%; // 确保占满父容器

    .captcha-input {
        flex: 1;
        margin-right: 8px;

        // 可选：如果希望输入框高度和验证码图片完全一致
        // :deep(.ant-input-affix-wrapper) {
        //     height: 38px; 
        // }
    }

    .captcha-img {
        width: 100px;
        height: 38px;
        cursor: pointer;
        border: 1px solid #d9d9d9;
        border-radius: 4px;
        // 防止图片被压缩变形
        object-fit: cover;
        background-color: #f5f5f5;
    }
}
</style>