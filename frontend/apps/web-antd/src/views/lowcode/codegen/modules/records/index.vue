<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { CodegenApi } from '#/api/lowcode/codegen';

import { useVbenModal } from '@vben/common-ui';

import { Button, message, Modal, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteGeneratedApi,
  deleteGenTableApi,
  getGenTableListApi,
} from '#/api/lowcode/codegen';
import { $t } from '#/locales';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import { useCodegenColumns, useFormOptions } from '../../data';

const emit = defineEmits<{
  /** 复制：把记录应用到设计器（由父组件处理页面切换） */
  useDesigner: [payload: { code: 'copy'; row: CodegenApi.GenTable }];
}>();

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useCodegenColumns(onActionClick),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: true,
      pageSize: DEFAULT_PAGE_SIZE,
      pageSizes: PAGE_SIZES,
    },
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          const res = await getGenTableListApi({
            ...page,
            ...formValues,
          });
          return res;
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
      search: true,
    },
  } as VxeTableGridOptions,
});

const [ModalApi, modalApi] = useVbenModal({
  onOpenChange(isOpen) {
    if (isOpen) {
      onRefresh();
    }
  },
});

function onRefresh() {
  gridApi.query();
}

function onActionClick({
  code,
  row,
}: {
  code: string;
  row: CodegenApi.GenTable;
}) {
  switch (code) {
    case 'copy': {
      // 带入设计器（组件2），由父组件处理页面切换
      emit('useDesigner', { code, row });
      break;
    }
    case 'delete': {
      // 删除记录：仅删除配置记录，不影响已生成的代码
      onDeleteRecord([row]);
      break;
    }
    case 'deleteModule': {
      // 删除整个模块：删除配置记录 + 生成的代码文件 + 真实数据表 + 落库菜单
      onDeleteModule([row]);
      break;
    }
  }
}

/** 删除记录：仅删除配置记录 */
function onDeleteRecord(rows: CodegenApi.GenTable[]) {
  Modal.confirm({
    title: $t('lowcode.codegen.confirm.deleteTitle'),
    content: $t('lowcode.codegen.confirm.deleteContent', [rows.length]),
    async onOk() {
      const ids = rows.map((item) => item.id);
      await deleteGenTableApi(ids);
      message.success($t('lowcode.codegen.message.deleteSuccess'));
      onRefresh();
    },
  });
}

/** 删除整个模块：删除配置记录 + 代码文件 + 真实数据表 + 落库菜单 */
function onDeleteModule(rows: CodegenApi.GenTable[]) {
  Modal.confirm({
    title: $t('lowcode.codegen.confirm.deleteModuleTitle'),
    content: $t('lowcode.codegen.confirm.deleteModuleContent', [rows.length]),
    okButtonProps: { danger: true },
    async onOk() {
      const ids = rows.map((item) => item.id);
      await deleteGeneratedApi(ids);
      message.success($t('lowcode.codegen.message.deleteModuleSuccess'));
      onRefresh();
    },
  });
}
</script>

<template>
  <ModalApi
    class="w-full max-w-[1100px]"
    :title="$t('lowcode.codegen.start.records.title')"
  >
    <template #footer>
      <div class="flex justify-end">
        <Button type="primary" @click="modalApi.close()">
          {{ $t('common.confirm') }}
        </Button>
      </div>
    </template>
    <Grid>
      <template #gen_type="{ row }">
        <Tag :color="row.genType === 'table' ? 'blue' : 'purple'">
          {{
            $t(
              `lowcode.codegen.gen_types.${row.genType === 'table' ? 'table' : 'form'}`,
            )
          }}
        </Tag>
      </template>
    </Grid>
  </ModalApi>
</template>
