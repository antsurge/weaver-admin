<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { InputSearch, Spin, Tree } from 'ant-design-vue';

import { getDepartmentTreeApi } from '#/api/organization/department';

export interface DeptNode {
  id: string;
  name: string;
  children?: DeptNode[];
}

const props = withDefaults(
  defineProps<{
    data?: DeptNode[];
    request?: () => Promise<DeptNode[]>;
    showSearch?: boolean;
    title?: string;
  }>(),
  {
    data: () => [],
    request: undefined,
    title: '',
    showSearch: true,
  },
);

const emit = defineEmits<{
  (e: 'select', node: DeptNode | null): void;
}>();

const searchValue = ref('');
const loading = ref(false);
const innerData = ref<DeptNode[]>([]);

const fieldNames = {
  title: 'name',
  key: 'id',
  children: 'children',
};

// 👉 默认请求（你可以替换成自己的接口）
async function defaultRequest(): Promise<DeptNode[]> {
  const res = await getDepartmentTreeApi();
  return res?.items ?? [];
}

async function loadData() {
  if (props.data && props.data.length > 0) {
    innerData.value = props.data;
    return;
  }

  loading.value = true;
  try {
    const fn = props.request || defaultRequest;
    innerData.value = await fn();
  } finally {
    loading.value = false;
  }
}

// 递归过滤
function filterTree(data: DeptNode[], searchValue: string): DeptNode[] {
  if (!searchValue) return data;

  return data
    .map((item) => {
      const children = item.children
        ? filterTree(item.children, searchValue)
        : [];

      if (
        item.name.includes(searchValue) ||
        (children && children.length > 0)
      ) {
        return { ...item, children };
      }

      return null;
    })
    .filter(Boolean) as DeptNode[];
}

const filteredData = computed(() => {
  return filterTree(innerData.value, searchValue.value);
});

function handleSelect(_: string[], e: any) {
  // 选中节点：emit 节点；取消选中：emit null，通知父组件清空过滤
  if (e?.selected) {
    emit('select', e.node);
  } else {
    emit('select', null);
  }
}

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="dept-tree">
    <!-- 搜索 -->
    <div v-if="showSearch" class="dept-tree__search">
      <InputSearch
        v-model:value="searchValue"
        placeholder="搜索部门"
        allow-clear
      />
    </div>
    <!-- 树 -->
    <Spin :spinning="loading">
      <Tree
        :tree-data="filteredData"
        :field-names="fieldNames"
        default-expand-all
        @select="handleSelect"
      />
    </Spin>
  </div>
</template>

<style scoped>
.dept-tree {
  height: 100%;
  padding: 12px;
  background: #fff;
}

.dept-tree__header {
  margin-bottom: 8px;
  font-weight: 500;
}

.dept-tree__search {
  margin-bottom: 8px;
}
</style>
