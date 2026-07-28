<script lang="ts" setup>
import { computed } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Button, Empty, Tag } from 'ant-design-vue';

import type { PermissionMenuApi } from '#/api/permission/menu';
import { $t } from '#/locales';

import ApiPermissionPickerModal from './api-permission-picker.vue';

const props = defineProps<{
  /** v-model 绑定值（数组） */
  value?: PermissionMenuApi.ApiPermission[];
  /** 字段 placeholder 文案 */
  placeholder?: string;
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
  const next = items.value.slice();
  next.splice(idx, 1);
  items.value = next;
  emit('change', next);
}

defineExpose({ openPicker });
</script>

<template>
  <div class="api-permission-picker flex flex-col gap-2">
    <div
      v-if="items.length > 0"
      class="rounded border border-gray-200 p-2 dark:border-gray-700"
    >
      <ul class="flex flex-col gap-1">
        <li
          v-for="(item, idx) in items"
          :key="`${item.service}-${item.method}-${item.path}`"
          class="flex items-center justify-between rounded bg-gray-50 px-2 py-1 text-xs dark:bg-gray-800"
        >
          <div class="flex items-center gap-2">
            <Tag color="blue">{{ item.method }}</Tag>
            <span class="font-mono">{{ item.path }}</span>
            <span class="text-gray-500">· {{ item.service }}</span>
            <span v-if="item.summary" class="text-gray-400">
              · {{ item.summary }}
            </span>
          </div>
          <Button
            type="link"
            size="small"
            class="!px-1"
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
    />
    <div>
      <Button type="dashed" block @click="openPicker">
        {{ $t('permission.menu.apiPermission.addButton') }}
      </Button>
    </div>

    <PickerModal @confirm="onPickerConfirm" />
  </div>
</template>