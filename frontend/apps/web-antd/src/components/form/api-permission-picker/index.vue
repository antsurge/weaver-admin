<script lang="ts" setup>
import type { PermissionMenuApi } from '#/api/permission/menu';

import { computed } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Button, Empty, Tag } from 'ant-design-vue';

import { $t } from '#/locales';

import ApiPermissionPickerModal from './api-permission-picker.vue';

const props = defineProps<{
  /** 字段 placeholder 文案 */
  placeholder?: string;
  /** v-model 绑定值（数组） */
  value?: PermissionMenuApi.ApiPermission[];
}>();

const emit = defineEmits<{
  (e: 'update:value', val: PermissionMenuApi.ApiPermission[]): void;
  (e: 'change', val: PermissionMenuApi.ApiPermission[]): void;
}>();

const items = computed<PermissionMenuApi.ApiPermission[]>({
  get: () => props.value ?? [],
  set: (v) => emit('update:value', v),
});

const [PickerModal, pickerModalApi] = useVbenModal({
  connectedComponent: ApiPermissionPickerModal,
  destroyOnClose: true,
});

function openPicker() {
  pickerModalApi.setData({ initialSelected: items.value }).open();
}

function onPickerConfirm(picked: PermissionMenuApi.ApiPermission[]) {
  items.value = picked;
  emit('change', picked);
}

function removeOne(idx: number) {
  const next = [...items.value];
  next.splice(idx, 1);
  items.value = next;
  emit('change', next);
}

defineExpose({ openPicker });
</script>

<template>
  <div class="api-permission-picker flex w-full flex-col gap-2">
    <div
      v-if="items.length > 0"
      class="w-full rounded border border-gray-200 p-2 dark:border-gray-700"
    >
      <ul class="flex w-full flex-col gap-1">
        <li
          v-for="(item, idx) in items"
          :key="`${item.service}-${item.method}-${item.path}`"
          class="flex w-full items-center justify-between rounded bg-gray-50 px-2 py-1 text-xs dark:bg-gray-800"
        >
          <div class="flex min-w-0 items-center gap-2">
            <Tag color="blue" class="!mr-0 shrink-0">{{ item.method }}</Tag>
            <span class="truncate font-mono">{{ item.path }}</span>
            <span class="shrink-0 text-gray-500">· {{ item.service }}</span>
            <span v-if="item.summary" class="truncate text-gray-400">
              · {{ item.summary }}
            </span>
          </div>
          <Button
            type="link"
            size="small"
            class="shrink-0 !px-1"
            @click="removeOne(idx)"
          >
            {{ $t('permission.menu.apiPermission.remove') }}
          </Button>
        </li>
      </ul>
    </div>
    <Empty
      v-else
      :description="placeholder ?? $t('permission.menu.apiPermission.empty')"
      class="w-full"
    />
    <Button type="dashed" block class="w-full" @click="openPicker">
      {{ $t('permission.menu.apiPermission.addButton') }}
    </Button>

    <PickerModal @confirm="onPickerConfirm" />
  </div>
</template>
