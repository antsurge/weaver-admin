<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { FormSchemaApi } from '#/api/lowcode/form-schema';

import { nextTick, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteFormSchemaApi,
  getFormSchemaListApi,
} from '#/api/lowcode/form-schema';
import { $t } from '#/locales';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import { useFormOptions, useFormSchemaColumns } from './data';
import FormDesigner from './modules/designer/index.vue';
import FormSchemaForm from './modules/form/index.vue';
import FormRender from './modules/render/index.vue';
import FormSubmissions from './modules/submissions/index.vue';

const [FormSchemaModal, formSchemaApi] = useVbenModal({
  connectedComponent: FormSchemaForm,
  destroyOnClose: true,
});

const [DesignerModal, designerApi] = useVbenModal({
  connectedComponent: FormDesigner,
  destroyOnClose: true,
});

const [RenderModal, renderApi] = useVbenModal({
  connectedComponent: FormRender,
  destroyOnClose: true,
});

const [SubmissionsModal, submissionsApi] = useVbenModal({
  connectedComponent: FormSubmissions,
  destroyOnClose: true,
});

const selectedRows = ref<FormSchemaApi.FormSchema[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useFormSchemaColumns(onActionClick),
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
          const res = await getFormSchemaListApi({
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

  gridEvents: {
    checkboxChange() {
      nextTick(() => {
        const records = (gridApi.grid as any)?.getCheckboxRecords?.() ?? [];
        selectedRows.value = records;
      });
    },
  },
});

function onRefresh() {
  gridApi.query();
}

function onCreate() {
  formSchemaApi.setData({}).open();
}

function onActionClick({
  code,
  row,
}: {
  code: string;
  row: FormSchemaApi.FormSchema;
}) {
  switch (code) {
    case 'delete': {
      onDelete([row]);
      break;
    }
    case 'design': {
      designerApi.setData(row).open();
      break;
    }
    case 'edit': {
      formSchemaApi.setData(row).open();
      break;
    }
    case 'preview': {
      renderApi.setData(row).open();
      break;
    }
    case 'submissions': {
      submissionsApi.setData(row).open();
      break;
    }
  }
}

function onDelete(rows: FormSchemaApi.FormSchema[]) {
  Modal.confirm({
    title: $t('lowcode.formBuilder.confirm.deleteTitle'),
    content: $t('lowcode.formBuilder.confirm.deleteContent', [rows.length]),
    async onOk() {
      const ids = rows.map((item) => item.id);
      await deleteFormSchemaApi(ids);
      message.success($t('lowcode.formBuilder.message.deleteSuccess'));
      onRefresh();
    },
  });
}

function onBatchDelete() {
  if (selectedRows.value.length === 0) return;
  onDelete(selectedRows.value);
}
</script>

<template>
  <Page auto-content-height>
    <Grid>
      <template #toolbar-actions>
        <Button type="primary" @click="onCreate">
          <Plus class="mr-1" />
          {{ $t('common.actions.create') }}
        </Button>
        <Button
          danger
          :disabled="selectedRows.length === 0"
          @click="onBatchDelete"
        >
          {{ $t('system.job.operation.batchDelete') }}
        </Button>
      </template>
      <template #status="{ row }">
        <Tag :color="row.status === 'enabled' ? 'success' : 'error'">
          {{
            row.status === 'enabled'
              ? $t('ui.status.enabled')
              : $t('ui.status.disabled')
          }}
        </Tag>
      </template>
    </Grid>
    <FormSchemaModal @success="onRefresh" />
    <DesignerModal @success="onRefresh" />
    <RenderModal @success="onRefresh" />
    <SubmissionsModal />
  </Page>
</template>
