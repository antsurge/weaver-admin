<!-- src/components/LanguageToggle.vue -->
<script setup lang="ts">
import { GlobalOutlined } from '@ant-design/icons-vue';
import { LOCALE_ARRAY } from "@/locales/constant"
import { useConfigStore } from "@/stores/config"
import { setI18nLanguage } from "@/locales";

const configStore = useConfigStore()

defineOptions({
  name: 'AuthLanguageToggle',
});

// 处理语言切换
// 处理语言切换
async function handleMenuClick({ key }: { key: string }) {
  if (key == configStore.lang.defaultLang) {
    return
  }
  setI18nLanguage(key)
}
</script>

<template>
  <a-dropdown trigger="click">
    <a-button type="text" class="flex-center p-0">
      <GlobalOutlined />
    </a-button>

    <template #overlay>
      <a-menu :selectedKeys="[configStore.lang.defaultLang]" @click="handleMenuClick">
        <a-menu-item v-for="item in LOCALE_ARRAY" :key="item.value">
          {{ item.label }}
        </a-menu-item>
      </a-menu>
    </template>
  </a-dropdown>
</template>

<style scoped lang="scss">
.flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

.a-button.flex-center {
  transition: transform 0.2s;
}

.a-button.flex-center:hover {
  transform: scale(1.05);
}
</style>