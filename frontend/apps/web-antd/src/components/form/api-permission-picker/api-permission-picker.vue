<script lang="ts" setup>
import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Checkbox, Empty, Input, Spin, Tag, Tree } from 'ant-design-vue';

import type { PermissionMenuApi } from '#/api/permission/menu';
import { listApiMetadataApi } from '#/api/permission/menu';
import { $t } from '#/locales';

const emit = defineEmits<{
  /** 确认选择 */
  confirm: [items: PermissionMenuApi.ApiPermission[]];
}>();

const props = defineProps<{
  /** 已选 API 权限（回显用） */
  initialSelected?: PermissionMenuApi.ApiPermission[];
}>();

// 元数据 + 加载状态
const metadata = ref<PermissionMenuApi.ApiMetadata[]>([]);
const loading = ref(false);
const expandedKeys = ref<string[]>([]);

// 已选 key 列表（用 "${service}::${method} ${path}" 区分不同 service）
const selectedKeys = ref<string[]>([]);
const selectedItems = ref<PermissionMenuApi.ApiPermission[]>([]);

// 搜索关键字
const searchKeyword = ref('');

function keyOf(api: { method: string; path: string }) {
  return `${api.method} ${api.path}`;
}

// 树形数据：service -> endpoints
const treeData = computed(() =>
  metadata.value.map((svc) => ({
    title: svc.service,
    key: svc.service,
    children: svc.endpoints.map((ep) => ({
      key: `${svc.service}::${ep.method} ${ep.path}`,
      title: `${ep.method} ${ep.path}`,
      isLeaf: true,
      dataRef: { service: svc.service, ...ep },
    })),
  })),
);

// 搜索过滤
const filteredTreeData = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase();
  if (!kw) return treeData.value;
  return treeData.value
    .map((node) => {
      const filteredChildren = (node.children || []).filter((child: any) => {
        const data = child.dataRef;
        return (
          data.path.toLowerCase().includes(kw) ||
          data.method.toLowerCase().includes(kw) ||
          (data.summary || '').toLowerCase().includes(kw) ||
          data.service.toLowerCase().includes(kw)
        );
      });
      if (filteredChildren.length === 0) return null;
      return { ...node, children: filteredChildren };
    })
    .filter(Boolean) as typeof treeData.value;
});

async function loadMetadata() {
  loading.value = true;
  try {
    metadata.value = await listApiMetadataApi();
    console.log('metadata', metadata.value)
    expandedKeys.value = metadata.value.map((s) => s.service);
  } finally {
    loading.value = false;
  }
}

function initSelected(items: PermissionMenuApi.ApiPermission[] = []) {
  selectedItems.value = items.map((it) => ({ ...it }));
  selectedKeys.value = items.map(
    (it) => `${it.service}::${keyOf(it)}`,
  );
}

function rebuildSelectedItems() {
  const next: PermissionMenuApi.ApiPermission[] = [];
  for (const k of selectedKeys.value) {
    const [svc, rest] = k.split('::');
    if (!svc || !rest) continue;
    const [method, ...pathParts] = rest.split(' ');
    const path = pathParts.join(' ');
    const ep = metadata.value
      .find((s) => s.service === svc)
      ?.endpoints.find((e) => e.method === method && e.path === path);
    if (ep) {
      next.push({
        service: svc,
        method: ep.method,
        path: ep.path,
        summary: ep.summary,
      });
    }
  }
  selectedItems.value = next;
}

function onCheck(checked: Array<string | number> | { checked: Array<string | number>; halfChecked: Array<string | number> }) {
  // ant-design-vue 的 onCheck 签名兼容两种形态：
  // - checkable 时回调为 (checkedKeys: Key[], info: CheckInfo)
  // - checkStrictly 时回调形如 (checkedKeys: Key[])
  let keys: Array<string | number>;
  if (Array.isArray(checked)) {
    keys = checked;
  } else {
    keys = checked.checked ?? [];
  }
  // 只保留叶子节点（接口）；父节点（service）只用作分组
  const leafKeys = keys.filter(
    (k): k is string => typeof k === 'string' && k.includes('::'),
  );
  selectedKeys.value = leafKeys;
  rebuildSelectedItems();
}

const [Modal] = useVbenModal({
  onOpenChange(isOpen: boolean) {
    if (!isOpen) return;
    void loadMetadata();
    initSelected(props.initialSelected);
  },
  onConfirm() {
    emit('confirm', selectedItems.value);
  },
});
</script>

<template>
  <Modal
    class="w-full max-w-[900px]"
    :title="$t('permission.menu.apiPermission.pickerTitle')"
  >
    <Spin :spinning="loading">
      <div class="flex gap-4" style="min-height: 480px">
        <!-- 左：服务树 -->
        <div class="w-1/2 flex flex-col gap-2">
          <Input
            v-model:value="searchKeyword"
            :placeholder="$t('permission.menu.apiPermission.searchPlaceholder')"
            allow-clear
          />
          <div
            class="flex-1 overflow-auto rounded border border-gray-200 p-2 dark:border-gray-700"
          >
            <Tree
              v-if="filteredTreeData.length > 0"
              v-model:expanded-keys="expandedKeys"
              :checkable="true"
              :tree-data="filteredTreeData"
              :checked-keys="selectedKeys"
              check-strictly
              default-expand-all
              @check="onCheck"
            >
              <template #title="{ dataRef }">
                <span v-if="dataRef.isLeaf" class="text-xs">
                  <Tag color="blue" class="mr-1">{{ dataRef.method }}</Tag>
                  <span class="font-mono">{{ dataRef.path }}</span>
                  <span v-if="dataRef.summary" class="ml-2 text-gray-500">
                    {{ dataRef.summary }}
                  </span>
                </span>
                <span v-else class="font-semibold">{{ dataRef.title }}</span>
              </template>
            </Tree>
            <Empty
              v-else
              :description="$t('permission.menu.apiPermission.empty')"
            />
          </div>
        </div>

        <!-- 右：已选列表 -->
        <div class="w-1/2 flex flex-col gap-2">
          <div class="text-sm font-medium">
            {{
              $t('permission.menu.apiPermission.selectedCount', [
                selectedItems.length,
              ])
            }}
          </div>
          <div
            class="flex-1 overflow-auto rounded border border-gray-200 p-2 dark:border-gray-700"
          >
            <Empty
              v-if="selectedItems.length === 0"
              :description="$t('permission.menu.apiPermission.selectedEmpty')"
            />
            <ul v-else class="flex flex-col gap-2">
              <li
                v-for="item in selectedItems"
                :key="`${item.service}-${item.method}-${item.path}`"
                class="flex items-center justify-between rounded border border-gray-100 px-2 py-1 dark:border-gray-700"
              >
                <div class="flex flex-col">
                  <div class="text-sm">
                    <Tag color="blue">{{ item.method }}</Tag>
                    <span class="font-mono">{{ item.path }}</span>
                  </div>
                  <div class="text-xs text-gray-500">
                    {{ item.service }}
                    <span v-if="item.summary"> · {{ item.summary }}</span>
                  </div>
                </div>
                <Checkbox
                  :checked="true"
                  @update:checked="
                    (v: boolean) => {
                      if (!v) {
                        const k = `${item.service}::${item.method} ${item.path}`;
                        selectedKeys = selectedKeys.filter((x) => x !== k);
                        rebuildSelectedItems();
                      }
                    }
                  "
                />
              </li>
            </ul>
          </div>
          <Alert
            type="info"
            show-icon
            :message="$t('permission.menu.apiPermission.tip')"
          />
        </div>
      </div>
    </Spin>
  </Modal>
</template>