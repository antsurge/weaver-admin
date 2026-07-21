<script lang="ts" setup>
import { computed, ref, onMounted } from 'vue';
import { getDepartmentTreeApi } from "#/api/organization/department"
import {
  Spin,
  InputSearch,
  Tree
} from 'ant-design-vue';

export interface DeptNode {
  id: string;
  name: string;
  children?: DeptNode[];
}

const props = withDefaults(defineProps<{
  title?: string;
  showSearch?: boolean;
  data?: DeptNode[];
  request?: () => Promise<DeptNode[]>;
}>(), {
  title: '',
  showSearch: true,
});

const emit = defineEmits<{
  (e: 'select', node: DeptNode): void;
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
  let res = await getDepartmentTreeApi()
  console.log("我是tree数据", res)
  return res?.items ?? [];
}

async function loadData() {
  if (props.data && props.data.length) {
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
  if (e.node) {
    emit('select', e.node);
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
      <InputSearch v-model:value="searchValue" placeholder="搜索部门" allow-clear />
    </div>
    <!-- 树 -->
    <Spin :spinning="loading">
      <Tree :tree-data="filteredData" :fieldNames="fieldNames" default-expand-all @select="handleSelect" />
    </Spin>
  </div>
</template>


<style scoped>
.dept-tree {
  padding: 12px;
  background: #fff;
  height: 100%;
}

.dept-tree__header {
  font-weight: 500;
  margin-bottom: 8px;
}

.dept-tree__search {
  margin-bottom: 8px;
}
</style>
