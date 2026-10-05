<script lang="ts" setup>
import type { CodegenApi } from '#/api/lowcode/codegen';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { createGenTableApi } from '#/api/lowcode/codegen';
import { $t } from '#/locales';

import CodegenDesigner from './modules/designer/index.vue';
import CodegenRecords from './modules/records/index.vue';
import CodegenStart from './modules/start/index.vue';
import TablePicker from './modules/table-picker/index.vue';

const [CodegenRecordsWrap, codegenRecordsApi] = useVbenModal({
  connectedComponent: CodegenRecords,
  destroyOnClose: true,
});

const [TablePickerWrap, tablePickerApi] = useVbenModal({
  connectedComponent: TablePicker,
  destroyOnClose: true,
});

// 视图模式：start=开始页，design=拖拽设计器（两大组件单页切换）
const viewMode = ref<'design' | 'start'>('start');

// 设计器状态：create=新建页面（可带默认数据），edit=编辑已有记录
const designerMode = ref<'create' | 'edit'>('create');
const designerData = ref<
  CodegenApi.GenTable | null | { table?: CodegenApi.DbTable }
>(null);
// 每次进入设计器递增，强制组件重新挂载以正确初始化
const designerKey = ref(0);

/** 新建（从零新建 CRUD）：进入拖拽设计器 create 模式 */
function onCreate() {
  designerMode.value = 'create';
  designerData.value = null;
  designerKey.value += 1;
  viewMode.value = 'design';
}

/** 选择数据表：先打开选表弹窗 */
function onPickTable() {
  tablePickerApi.setData({});
  tablePickerApi.open();
}

/** 选定数据表后：进入新建页面（create 模式），把所选表作为默认数据传入设计器 */
function onTableSelected(table: CodegenApi.DbTable) {
  designerMode.value = 'create';
  designerData.value = { table };
  designerKey.value += 1;
  viewMode.value = 'design';
}

/** CRUD 记录：打开表格弹框 */
function onShowRecords() {
  codegenRecordsApi.setData({});
  codegenRecordsApi.open();
}

/**
 * 来自 CRUD 记录弹框的"复制"操作：复制该记录（连同数据表字段）进入设计器（组件2）
 */
function onUseDesigner(payload: { code: 'copy'; row: CodegenApi.GenTable }) {
  codegenRecordsApi.close();
  void onCopy(payload.row);
}

/** 复制 CRUD 记录：基于既有配置创建一份新配置（表名带 _copy 后缀），随后进入设计器调整 */
async function onCopy(row: CodegenApi.GenTable) {
  try {
    const newTableName = `${row.tableName}_copy`;
    const payload = {
      id: '',
      tableName: newTableName,
      tableComment: `${row.tableComment || row.tableName}（副本）`,
      moduleName: row.moduleName || 'system',
      bizName: row.bizName || '',
      genType: row.genType || 'table',
      status: 'enabled',
      fieldsJson: row.fieldsJson || { fields: [] },
    };
    const created = await createGenTableApi(payload as any);
    message.success($t('lowcode.codegen.message.copySuccess'));
    const id = (created as any)?.id;
    if (id) {
      // 复制成功后直接进入设计器调整字段
      designerMode.value = 'edit';
      designerData.value = { ...payload, id };
      designerKey.value += 1;
      viewMode.value = 'design';
    }
  } catch {
    message.error($t('lowcode.codegen.message.copyFail'));
  }
}

/** 设计器生成完成：停留当前页，由用户决定继续或返回 */
function onDesignerSuccess(_id: string) {
  // no-op
}

/** 设计器放弃/取消：返回开始页 */
function onDesignerCancel() {
  viewMode.value = 'start';
  designerData.value = null;
}
</script>

<template>
  <Page auto-content-height>
    <!-- ============ 组件1：开始页 ============ -->
    <CodegenStart
      v-if="viewMode === 'start'"
      @create="onCreate"
      @table="onPickTable"
      @records="onShowRecords"
    />

    <!-- ============ 组件2：拖拽设计器 ============ -->
    <CodegenDesigner
      v-else
      :key="designerKey"
      class="h-full"
      :mode="designerMode"
      :initial-data="designerData"
      @cancel="onDesignerCancel"
      @success="onDesignerSuccess"
    />

    <!-- ============ 弹窗：CRUD 记录表格 ============ -->
    <CodegenRecordsWrap @use-designer="onUseDesigner" />

    <TablePickerWrap @select="onTableSelected" />
  </Page>
</template>
