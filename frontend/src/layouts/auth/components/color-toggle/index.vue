<!-- src/components/AuthColorToggle.vue -->
<script setup lang="ts">
import type { BuiltinThemeType } from "@/types"
import { BgColorsOutlined } from '@ant-design/icons-vue'

import {
    COLOR_PRESETS,
    preferences,
    updatePreferences,
} from '@/preferences'

defineOptions({
    name: 'AuthColorToggle'
})

// 点击切换主题
function handleUpdate(colorPrimary: string, type: BuiltinThemeType) {
    updatePreferences({
        theme: {
            colorPrimary,
            builtinType: type,
        },
    });
}
</script>

<template>
  <div class="group relative flex items-center overflow-hidden">
    <!-- 主题色预设列表 -->
    <div class="flex w-0 overflow-hidden transition-all duration-500 ease-out group-hover:w-66">
      <template v-for="preset in COLOR_PRESETS" :key="preset.color">
        <a-button
          type="text"
          class="flex-center p-0"
          @click="handleUpdate(preset.color, preset.type)"
        >
          <div
            :style="{ backgroundColor: preset.color }"
            class="flex-center relative size-5 rounded-full hover:scale-110"
          >
            <!-- 当前主题选中标记 -->
            <svg
              v-if="preferences.theme.builtinType === preset.type"
              class="h-3.5 w-3.5 text-white"
              height="1em"
              viewBox="0 0 15 15"
              width="1em"
            >
              <path
                clip-rule="evenodd"
                d="M11.467 3.727c.289.189.37.576.181.865l-4.25 6.5a.625.625 0 0 1-.944.12l-2.75-2.5a.625.625 0 0 1 .841-.925l2.208 2.007l3.849-5.886a.625.625 0 0 1 .865-.181"
                fill="currentColor"
                fill-rule="evenodd"
              />
            </svg>
          </div>
        </a-button>
      </template>
    </div>

    <!-- 主切换按钮 -->
    <a-button type="text" class="flex-center p-0">
      <BgColorsOutlined />
    </a-button>
  </div>
</template>

<style scoped lang="scss">
.flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 保持 hover 放大效果 */
.a-button.flex-center {
  transition: transform 0.2s;
}
.a-button.flex-center:hover {
  transform: scale(1.1);
}

/* 圆点大小 */
.size-5 {
  width: 1.25rem; /* 20px */
  height: 1.25rem;
}

/* 圆点 hover 缩放 */
.hover\:scale-110:hover {
  transform: scale(1.1);
}
</style>