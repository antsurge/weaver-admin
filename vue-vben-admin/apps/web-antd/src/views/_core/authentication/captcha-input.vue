<script setup lang="ts">
import { ref, watch,onMounted } from 'vue'
import { Input } from 'ant-design-vue'
import { getCaptchaApi } from '#/api'

const props = defineProps<{
  // 接收触发器
  refreshTrigger?: number
}>()

const emit = defineEmits(['update:modelValue', 'update:captchaId'])

const value = ref('')
const captchaImg = ref('')
const captchaId = ref('')

// 加载验证码
async function loadCaptcha() {
  const resData = await getCaptchaApi()
  captchaImg.value = resData.imageBase64
  captchaId.value = resData.captchaId

  emit('update:modelValue', value.value)
  emit('update:captchaId', captchaId.value)
}

// 输入更新
function updateValue(e: any) {
  value.value = e.target.value
  emit('update:modelValue', value.value)
}

function getCaptchaId() {
  return captchaId.value
}

// 暴露方法给父组件
defineExpose({
  refresh: loadCaptcha,
  getCaptchaId
})

// 4. 监听触发器的变化
watch(() => props.refreshTrigger, () => {
  if (props.refreshTrigger && props.refreshTrigger > 0) {
    loadCaptcha();
  }
});

// 如果父组件传了 ref，就把 refresh 方法挂上去
// 关键：在挂载时将方法暴露给父组件传入的 ref
onMounted(() => {
  // console.log(2222,props.captchaControl)
  // 关键：将子组件的方法“注册”回父组件的对象中
  // if (props.captchaControl) {
  //   props.captchaControl.value.refresh = loadCaptcha;
  // }
  loadCaptcha()
})
</script>

<template>
  <div class="captcha-wrapper">
    <Input
      :value="value"
      placeholder="请输入验证码"
      @input="updateValue"
    />
    <img
      class="captcha-img"
      :src="captchaImg"
      @click="loadCaptcha"
    />
  </div>
</template>

<style scoped>
.captcha-wrapper {
  display: flex;
  gap: 8px;
  width: 100%;
}

.captcha-wrapper :deep(.ant-input) {
  height: 40px;
}

.captcha-img {
  width: 110px;
  height: 40px;
  cursor: pointer;
  border-radius: 6px;
  border: 1px solid #d9d9d9;
}
</style>
