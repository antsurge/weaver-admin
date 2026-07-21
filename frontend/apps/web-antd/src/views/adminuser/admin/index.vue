<script lang="ts" setup>
import { reactive, ref } from 'vue';

import { ColPage } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import Table from "./modules/table/index.vue"
import DepartmentTree from "#/components/department-tree/index.vue"

import {
  Button,
  Card,
  Checkbox,
  Slider,
  Tooltip,
} from 'ant-design-vue';

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
const leftMinWidth = ref(props.leftMinWidth || 1);
const leftMaxWidth = ref(props.leftMaxWidth || 100);
</script>
<template>
  <ColPage
    auto-content-height
    description=""
    v-bind="props"
    title=""
  >
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
        class="mr-2 rounded-(--radius) border border-border bg-card p-2"
      >
        <DepartmentTree @select="(value:any)=>{console.log('我是选中的数据',value)}"/>
      </div>
    </template>
    <Table />
  </ColPage>
</template>
