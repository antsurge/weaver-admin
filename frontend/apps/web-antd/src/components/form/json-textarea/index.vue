<script lang="ts" setup>
import { ref, watch } from 'vue';

import { IconifyIcon } from '@vben/icons';
import { $t } from '@vben/locales';

import { Button, Input, message, Tooltip } from 'ant-design-vue';

const props = defineProps<{
  /** 是否禁用 */
  disabled?: boolean;
  /** 占位文本 */
  placeholder?: string;
  /** v-model 绑定值（JSON 字符串） */
  value?: string;
}>();

const emit = defineEmits<{
  (e: 'update:value', val: string): void;
}>();

const text = ref(props.value ?? '');

watch(
  () => props.value,
  (v) => {
    text.value = v ?? '';
  },
);

function handleChange(e: { target: { value: string } }) {
  const val = e.target.value;
  text.value = val;
  emit('update:value', val);
}

function formatJson() {
  if (!text.value?.trim()) return;
  try {
    const parsed = JSON.parse(text.value);
    text.value = JSON.stringify(parsed, null, 2);
    emit('update:value', text.value);
    message.success($t('ui.jsonViewer.formatSuccess'));
  } catch {
    message.error($t('ui.jsonViewer.formatError'));
  }
}
</script>

<template>
  <div class="json-textarea relative w-full">
    <Input.Textarea
      :value="text"
      :placeholder="placeholder"
      :disabled="disabled"
      :auto-size="{ minRows: 3, maxRows: 10 }"
      class="pr-16 font-mono text-xs"
      @change="handleChange"
    />
    <Tooltip :title="$t('ui.jsonViewer.format')">
      <Button
        class="absolute right-2 top-2 z-10"
        size="small"
        type="text"
        :disabled="disabled"
        @click="formatJson"
      >
        <IconifyIcon icon="ant-design:code-outlined" class="size-4" />
      </Button>
    </Tooltip>
  </div>
</template>
