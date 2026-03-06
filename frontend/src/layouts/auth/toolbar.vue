<script setup lang="ts">
import type { ToolbarType } from "./types"

import { computed } from "vue"

import {
  AuthColorToggle,
  AuthLanguageToggle,
  AuthThemeToggle,
} from "./components"

interface Props {
  toolbarList?: ToolbarType[]
}

defineOptions({
  name: 'AuthToolbar'
});

const props = withDefaults(defineProps<Props>(), {
  toolbarList: () => ['color', 'language', 'layout', 'theme'],
});

const showColor = computed(() => props.toolbarList.includes('color'));
const showLayout = computed(() => props.toolbarList.includes('layout'));
const showLanguage = computed(() => props.toolbarList.includes('language'));
const showTheme = computed(() => props.toolbarList.includes('theme'));

</script>
<template>
  <div :class="{
    'bg-accent rounded-3xl px-3 py-1': toolbarList.length > 1,
  }" class="flex flex-center absolute right-2 top-4 z-10">
    <div class="hidden md:flex">
      <AuthColorToggle v-if="showColor" />
    </div>
    <AuthLanguageToggle v-if="showLanguage" />
    <!-- <AuthThemeToggle v-if="showTheme"/> -->
  </div>
</template>
<style scoped lang="scss">
.bg-accent {
  background-color: hsl(240 5% 96%);
}
</style>