<script lang="ts" setup>
import type { PermissionMenuApi } from '#/api/permission/menu';
import type { SystemApiInterfaceApi } from '#/api/system/api-interface';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Checkbox, Empty, Input, Spin, Tag, Tree } from 'ant-design-vue';

import { getApiInterfaceListApi } from '#/api/system/api-interface';
import { $t } from '#/locales';

const props = defineProps<{
  /** 已选 API 权限（回显用） */
  initialSelected?: PermissionMenuApi.ApiPermission[];
}>();

const emit = defineEmits<{
  /** 确认选择 */
  confirm: [items: PermissionMenuApi.ApiPermission[]];
}>();

// 接口列表 + 加载状态
const interfaces = ref<SystemApiInterfaceApi.ApiInterface[]>([]);
const loading = ref(false);
const expandedKeys = ref<string[]>([]);

// 已选 key 列表（用 "tag::METHOD /path" 唯一标识接口）
const selectedKeys = ref<string[]>([]);
const selectedItems = ref<PermissionMenuApi.ApiPermission[]>([]);

// 搜索关键字
const searchKeyword = ref('');

function tagOf(item: { service?: string; tag?: string }) {
  return item.tag || item.service || '未分类';
}

/** 接口唯一 key */
function itemKey(item: {
  method: string;
  path: string;
  service?: string;
  tag?: string;
}) {
  return `${tagOf(item)}::${item.method} ${item.path}`;
}

// 树形数据：tag -> 接口（两级）
const treeData = computed(() => {
  const tagMap = new Map<string, SystemApiInterfaceApi.ApiInterface[]>();
  for (const item of interfaces.value) {
    const tag = tagOf(item);
    if (!tagMap.has(tag)) tagMap.set(tag, []);
    tagMap.get(tag)!.push(item);
  }
  return [...tagMap.entries()].map(([tag, items]) => ({
    title: `${tag} (${items.length})`,
    key: `tag::${tag}`,
    // 分组节点只作展示，不可勾选
    selectable: false,
    checkable: false,
    children: items.map((item) => ({
      key: itemKey(item),
      title: `${item.method} ${item.path}`,
      isLeaf: true,

      apiRef: item,
    })),
  }));
});

// 已选列表按 tag 分组
const selectedGroups = computed(() => {
  const tagMap = new Map<string, PermissionMenuApi.ApiPermission[]>();
  for (const item of selectedItems.value) {
    const tag = tagOf(item);
    if (!tagMap.has(tag)) tagMap.set(tag, []);
    tagMap.get(tag)!.push(item);
  }
  return [...tagMap.entries()].map(([tag, items]) => ({ tag, items }));
});

// 搜索过滤（tag -> 接口）
const filteredTreeData = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase();
  if (!kw) return treeData.value;
  return treeData.value
    .map((tagNode) => {
      // tag 名命中时保留该 tag 下全部接口
      if (tagNode.title.toLowerCase().includes(kw)) return tagNode;
      const children = (tagNode.children || []).filter((ep: any) => {
        const data = ep.apiRef;
        return (
          data.path.toLowerCase().includes(kw) ||
          data.method.toLowerCase().includes(kw) ||
          (data.summary || '').toLowerCase().includes(kw) ||
          (data.service || '').toLowerCase().includes(kw)
        );
      });
      if (children.length === 0) return null;
      return { ...tagNode, children };
    })
    .filter(Boolean) as typeof treeData.value;
});

// 搜索时自动展开命中的分组
const effectiveExpandedKeys = computed(() =>
  searchKeyword.value.trim()
    ? filteredTreeData.value.map((tagNode) => tagNode.key)
    : expandedKeys.value,
);

/**
 * 传给 Tree 的受控勾选 key。
 * 搜索过滤后 treeData 会丢掉未命中的接口，若 checkedKeys 仍引用这些
 * 不存在的 key，ant-design-vue(vc-tree) 的 getTreeNodeProps 会因
 * entity 缺失访问 entity.parent 抛 TypeError（known issue #7549）。
 * 因此这里只保留当前过滤结果里真实存在的 key。
 */
const effectiveCheckedKeys = computed(() => {
  const valid = new Set<string>();
  const collect = (nodes: any[]) => {
    for (const n of nodes) {
      valid.add(n.key as string);
      if (n.children) collect(n.children);
    }
  };
  collect(filteredTreeData.value);
  return selectedKeys.value.filter((k) => valid.has(k));
});

async function loadInterfaces() {
  loading.value = true;
  try {
    // 直接取「接口管理」的数据（pageSize=0 表示取全部）
    const res = await getApiInterfaceListApi({ pageSize: 0 });
    interfaces.value = (res?.items ?? []).filter(
      (it) => it.service && it.method && it.path,
    );
    // 默认展开所有 tag 分组
    expandedKeys.value = [
      ...new Set(interfaces.value.map((it) => `tag::${tagOf(it)}`)),
    ];
  } catch (error) {
    console.error('[api-interface] load failed =>', error);
    interfaces.value = [];
  } finally {
    loading.value = false;
  }
}

function initSelected(items: PermissionMenuApi.ApiPermission[] = []) {
  selectedItems.value = items.map((it) => ({ ...it }));
  selectedKeys.value = items.map((it) => itemKey(it));
}

function rebuildSelectedItems() {
  const next: PermissionMenuApi.ApiPermission[] = [];
  for (const k of selectedKeys.value) {
    // 直接按 key（tag::METHOD path）精确匹配接口列表中的叶子节点
    const item = interfaces.value.find((it) => itemKey(it) === k);
    if (item) {
      next.push({
        service: item.service,
        tag: item.tag,
        method: item.method,
        path: item.path,
        summary: item.summary,
        code: item.code,
      });
    }
  }
  selectedItems.value = next;
}

/** 从已选列表移除单个接口 */
function removeSelected(item: PermissionMenuApi.ApiPermission) {
  const k = itemKey(item);
  selectedKeys.value = selectedKeys.value.filter((x) => x !== k);
  rebuildSelectedItems();
}

/** 收集当前过滤结果树中「可见」的叶子接口 key（隐藏的在过滤之外，不会被 Tree 回调） */
function collectVisibleLeafKeys() {
  const visible = new Set<string>();
  const walk = (
    nodes: Array<{ apiRef?: unknown; children?: any[]; key?: string }>,
  ) => {
    for (const n of nodes) {
      if (n.apiRef) visible.add(n.key as string);
      if (n.children) walk(n.children);
    }
  };
  walk(filteredTreeData.value as any);
  return visible;
}

function onCheck(
  checked:
    | Array<number | string>
    | { checked: Array<number | string>; halfChecked: Array<number | string> },
) {
  // ant-design-vue 的 onCheck 签名兼容两种形态：
  // - checkable 时回调为 (checkedKeys: Key[], info: CheckInfo)
  // - checkStrictly 时回调形如 (checkedKeys: Key[])
  const keys: Array<number | string> = Array.isArray(checked)
    ? checked
    : (checked.checked ?? []);
  // 只保留叶子节点（接口）；tag 分组节点不可勾选
  const leafKeys = keys.filter(
    (k): k is string => typeof k === 'string' && !k.startsWith('tag::'),
  );
  // 关键：合并而非覆盖。搜索过滤时 Tree 只会基于可见树回调勾选 key，
  // 若直接赋值会给已选列表清掉隐藏在过滤结果之外的接口。
  const visible = collectVisibleLeafKeys();
  const kept = selectedKeys.value.filter((k) => !visible.has(k));
  selectedKeys.value = [...new Set([...kept, ...leafKeys])];
  rebuildSelectedItems();
}

const [Modal, modalApi] = useVbenModal({
  onOpenChange(isOpen: boolean) {
    if (!isOpen) return;
    // connectedComponent 模式下，外层通过 pickerModalApi.setData() 传参，
    // 需用 modalApi.getData() 读取（props 收不到 setData 的值）
    const data = modalApi.getData<{
      initialSelected?: PermissionMenuApi.ApiPermission[];
    }>();
    void loadInterfaces();
    initSelected(data?.initialSelected ?? props.initialSelected ?? []);
  },
  onConfirm() {
    emit('confirm', selectedItems.value);
    modalApi.close();
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
        <!-- 左：接口树（按 tag 分组） -->
        <div class="flex w-1/2 flex-col gap-2">
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
              :key="searchKeyword"
              :expanded-keys="effectiveExpandedKeys"
              :checkable="true"
              :tree-data="filteredTreeData"
              :checked-keys="effectiveCheckedKeys"
              check-strictly
              @check="onCheck"
              @expand="
                (keys: any) => {
                  if (!searchKeyword.trim()) expandedKeys = keys;
                }
              "
            >
              <template #title="{ apiRef, title }">
                <!-- 叶子节点：接口（key 形如 "tag::GET /path"） -->
                <span v-if="apiRef" class="text-xs">
                  <Tag color="blue" class="mr-1">{{ apiRef.method }}</Tag>
                  <span class="font-mono">{{ apiRef.path }}</span>
                  <span v-if="apiRef.summary" class="ml-2 text-gray-500">
                    {{ apiRef.summary }}
                  </span>
                </span>
                <!-- tag 分组节点 -->
                <span
                  v-else
                  class="font-semibold text-blue-600 dark:text-blue-400"
                >
                  {{ title }}
                </span>
              </template>
            </Tree>
            <Empty
              v-else
              :description="$t('permission.menu.apiPermission.empty')"
            />
          </div>
        </div>

        <!-- 右：已选列表 -->
        <div class="flex w-1/2 flex-col gap-2">
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
            <div v-else class="flex flex-col gap-3">
              <div v-for="group in selectedGroups" :key="group.tag">
                <!-- tag 分组标题 -->
                <div
                  class="mb-1 flex items-center gap-1 text-sm font-semibold text-blue-600 dark:text-blue-400"
                >
                  <span>{{ group.tag }}</span>
                  <span class="text-xs text-gray-400">
                    ({{ group.items.length }})
                  </span>
                </div>
                <ul class="flex flex-col gap-1">
                  <li
                    v-for="item in group.items"
                    :key="`${item.service}-${item.method}-${item.path}`"
                    class="flex items-center justify-between rounded border border-gray-100 px-2 py-1 dark:border-gray-700"
                  >
                    <div class="flex flex-col">
                      <div class="text-sm">
                        <Tag color="blue">{{ item.method }}</Tag>
                        <span class="font-mono">{{ item.path }}</span>
                      </div>
                      <div v-if="item.summary" class="text-xs text-gray-500">
                        {{ item.summary }}
                      </div>
                    </div>
                    <Checkbox
                      :checked="true"
                      @update:checked="
                        (v: boolean) => {
                          if (!v) removeSelected(item);
                        }
                      "
                    />
                  </li>
                </ul>
              </div>
            </div>
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
