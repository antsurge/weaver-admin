<!-- src/components/LayoutHeader.vue -->
<template>
    <a-layout-header class="flex items-center justify-between px-4 h-16 bg-white shadow-sm">
      <!-- 左侧插槽 -->
      <div class="flex items-center space-x-2">
        <template v-for="slot in leftSlots.filter(item => item.index < REFERENCE_VALUE)" :key="slot.name">
          <slot :name="slot.name">
            <template v-if="slot.name === 'refresh'">
              <a-button type="text" icon="ReloadOutlined" @click="refresh" />
            </template>
          </slot>
        </template>
      </div>
  
      <!-- 中间面包屑 -->
      <div class="flex-1 flex justify-center">
        <slot name="breadcrumb" />
      </div>
  
      <!-- 右侧插槽 -->
      <div class="flex items-center space-x-2">
        <template v-for="slot in rightSlots" :key="slot.name">
          <slot :name="slot.name">
            <template v-if="slot.name === 'global-search'">
              <GlobalSearch
                :enable-shortcut-key="globalSearchShortcutKey"
                :menus="accessStore.accessMenus"
                class="mr-2"
              />
            </template>
            <template v-else-if="slot.name === 'preferences'">
              <PreferencesButton @clear-preferences-and-logout="clearPreferencesAndLogout" />
            </template>
            <template v-else-if="slot.name === 'theme-toggle'">
              <ThemeToggle />
            </template>
            <template v-else-if="slot.name === 'language-toggle'">
              <LanguageToggle />
            </template>
            <template v-else-if="slot.name === 'fullscreen'">
              <VbenFullScreen />
            </template>
            <template v-else-if="slot.name === 'timezone'">
              <TimezoneButton />
            </template>
          </slot>
        </template>
      </div>
    </a-layout-header>
  </template>
  
  <script lang="ts" setup>
  import { ref, computed, useSlots } from 'vue';
  import { useRefresh } from '@vben/hooks';
  import { preferences, usePreferences } from '@vben/preferences';
  import { useAccessStore } from '@vben/stores';
  
  import { GlobalSearch, LanguageToggle, PreferencesButton, ThemeToggle, TimezoneButton } from '../../widgets';
  import { VbenFullScreen } from '@vben-core/shadcn-ui';
  import { ReloadOutlined } from '@ant-design/icons-vue';
  
  interface Props {
    theme?: 'light' | 'dark';
  }
  
  defineOptions({
    name: 'LayoutHeader',
  });
  
  withDefaults(defineProps<Props>(), {
    theme: 'light',
  });
  
  const emit = defineEmits<{ clearPreferencesAndLogout: [] }>();
  
  const REFERENCE_VALUE = 50;
  
  const accessStore = useAccessStore();
  const { globalSearchShortcutKey, preferencesButtonPosition } = usePreferences();
  const slots = useSlots();
  const { refresh } = useRefresh();
  
  // 左侧插槽
  const leftSlots = computed(() => {
    const list: Array<{ index: number; name: string }> = [];
    if (preferences.widget.refresh) {
      list.push({ index: 0, name: 'refresh' });
    }
  
    Object.keys(slots).forEach(key => {
      const name = key.split('-');
      if (key.startsWith('header-left')) {
        list.push({ index: Number(name[2]), name: key });
      }
    });
    return list.sort((a, b) => a.index - b.index);
  });
  
  // 右侧插槽
  const rightSlots = computed(() => {
    const list: Array<{ index: number; name: string }> = [{ index: REFERENCE_VALUE + 100, name: 'user-dropdown' }];
  
    if (preferences.widget.globalSearch) list.push({ index: REFERENCE_VALUE, name: 'global-search' });
    if (preferencesButtonPosition.value.header) list.push({ index: REFERENCE_VALUE + 10, name: 'preferences' });
    if (preferences.widget.themeToggle) list.push({ index: REFERENCE_VALUE + 20, name: 'theme-toggle' });
    if (preferences.widget.languageToggle) list.push({ index: REFERENCE_VALUE + 30, name: 'language-toggle' });
    if (preferences.widget.timezone) list.push({ index: REFERENCE_VALUE + 40, name: 'timezone' });
    if (preferences.widget.fullscreen) list.push({ index: REFERENCE_VALUE + 50, name: 'fullscreen' });
    if (preferences.widget.notification) list.push({ index: REFERENCE_VALUE + 60, name: 'notification' });
  
    Object.keys(slots).forEach(key => {
      const name = key.split('-');
      if (key.startsWith('header-right')) {
        list.push({ index: Number(name[2]), name: key });
      }
    });
  
    return list.sort((a, b) => a.index - b.index);
  });
  
  function clearPreferencesAndLogout() {
    emit('clearPreferencesAndLogout');
  }
  </script>
  
  <style scoped>
  /* 可选 AntD 风格自定义 */
  </style>