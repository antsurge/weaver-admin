<script lang="ts" setup>
import type { DeptNode } from '#/components/department-tree/index.vue';

import { reactive, ref } from 'vue';

import { ColPage } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import { Button, Tooltip } from 'ant-design-vue';

import DepartmentTree from '#/components/department-tree/index.vue';

import Table from './modules/table/index.vue';

const props = reactive({
  leftCollapsedWidth: 5,
  leftCollapsible: true,
  leftMaxWidth: 30,
  leftMinWidth: 20,
  leftWidth: 10,
  resizable: true,
  rightWidth: 70,
  splitHandle: true,
  splitLine: true,
});

// 左侧选中的部门（用于列表过滤 + 新增时默认归属）
const selectedDept = ref<DeptNode | null>(null);

function onSelectDept(node: DeptNode | null) {
  selectedDept.value = node;
}
</script>
<template>
  <ColPage auto-content-height description="" v-bind="props" title="">
    <template #left="{ isCollapsed, expand }">
      <div v-if="isCollapsed" @click="expand">
        <Tooltip title="点击展开左侧">
          <Button shape="circle" type="primary" class="flex-center">
            <template #icon>
              <IconifyIcon class="text-2xl" icon="bi:arrow-right" />
            </template>
          </Button>
        </Tooltip>
      </div>
      <div
        v-else
        :style="{ minWidth: '120px' }"
        class="rounded-(--radius) mr-2 border border-border bg-card p-2"
      >
        <DepartmentTree @select="onSelectDept" />
      </div>
    </template>
    <Table
      :department-id="selectedDept?.id"
      :department-name="selectedDept?.name"
    />
  </ColPage>
</template>
