<script lang="ts" setup>
import type { ThemeModeType } from '@vben/types';

import { computed } from 'vue';
import { Tooltip, Segmented } from 'ant-design-vue';
import {
  BulbOutlined,
  MoonOutlined,
  DesktopOutlined,
} from '@ant-design/icons-vue';

import { $t } from '@/locales/index1';
import {
  preferences,
  updatePreferences,
} from '@/preferences';

defineOptions({
  name: 'ThemeToggle',
});

withDefaults(defineProps<{ shouldOnHover?: boolean }>(), {
  shouldOnHover: false,
});

const value = computed({
  get: () => preferences.theme.mode,
  set: (val: ThemeModeType) => {
    updatePreferences({
      theme: { mode: val },
    });
  },
});

const options = [
  {
    label: h(BulbOutlined),
    value: 'light',
  },
  {
    label: h(MoonOutlined),
    value: 'dark',
  },
  {
    label: h(DesktopOutlined),
    value: 'auto',
  },
];
</script>

<template>
  <Tooltip
    :title="$t('preferences.theme.title')"
    :open="shouldOnHover ? undefined : false"
    placement="bottom"
  >
    <Segmented
      v-model:value="value"
      :options="options"
      size="middle"
    />
  </Tooltip>
</template>