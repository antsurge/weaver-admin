<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { Order } from '#/api/order/order';

import { reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, Form, FormItem, Input, message, Modal } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  createOrderApi,
  deleteOrderApi,
  listOrderApi,
  updateOrderApi,
} from '#/api/order/order';
import { $t } from '#/locales';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

defineOptions({ name: 'OrderManagement' });

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [],
  },
  gridOptions: {
    columns: [
      {
        field: 'name',
        align: 'center',
        title: 'name',
        minWidth: 150,
      },
      {
        field: 'sn',
        align: 'center',
        title: 'sn',
        minWidth: 150,
      },
      {
        field: 'operation',
        align: 'center',
        title: $t('common.fields.operation'),
        width: 150,
        slots: { default: 'operation' },
      },
    ],
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
          const res = await listOrderApi({
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
    checkboxConfig: {
      checkStrictly: true,
      showHeader: true,
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
const selectedRows = ref<Order[]>([]);

const formRef = ref();
const modalVisible = ref(false);
const submitting = ref(false);
const isEdit = ref(false);
const form = reactive<Partial<Order>>({
  name: undefined,
  sn: undefined,
});

const rules: Record<string, any> = {};

function onRefresh() {
  gridApi.query();
}
function onCheckboxChange() {
  const checkboxRecords = gridApi.grid?.getCheckboxRecords?.() ?? [];
  selectedRows.value = checkboxRecords;
}

function onBatchDelete() {
  if (selectedRows.value.length === 0) return;
  const ids = selectedRows.value.map((r) => r.id);
  Modal.confirm({
    title: $t('common.confirm'),
    content: $t('common.actions.delete'),
    async onOk() {
      await deleteOrderApi(ids);
      message.success('删除成功');
      selectedRows.value = [];
      onRefresh();
    },
  });
}

function onCreate() {
  isEdit.value = false;
  Object.keys(form).forEach((key) => {
    form[key as keyof typeof form] = undefined;
  });
  modalVisible.value = true;
}

function onEdit(row: Order) {
  isEdit.value = true;
  Object.assign(form, row);
  modalVisible.value = true;
}

function onDelete(row: Order) {
  Modal.confirm({
    title: $t('common.confirm'),
    content: $t('common.actions.delete'),
    async onOk() {
      await deleteOrderApi([row.id]);
      message.success('删除成功');
      onRefresh();
    },
  });
}

async function onSubmit() {
  try {
    await formRef.value.validate();
  } catch {
    return;
  }
  submitting.value = true;
  try {
    await (isEdit.value
      ? updateOrderApi(form.id as string, form)
      : createOrderApi(form));
    message.success(isEdit.value ? '更新成功' : '创建成功');
    modalVisible.value = false;
    onRefresh();
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <Page auto-content-height>
    <Grid @checkbox-change="onCheckboxChange" @checkbox-all="onCheckboxChange">
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            v-access:code="['Order:Order:Create']"
            type="primary"
            @click="onCreate"
          >
            <Plus class="mr-1" />
            <span v-text="$t('common.actions.create')"></span>
          </Button>
          <Button
            v-access:code="['Order:Order:BatchDelete']"
            danger
            :disabled="selectedRows.length === 0"
            @click="onBatchDelete"
          >
            <span v-text="$t('common.actions.delete')"></span>
          </Button>
        </div>
      </template>
      <template #operation="{ row }">
        <Button type="link" @click="onEdit(row)">
          {{ $t('common.actions.edit') }}
        </Button>
        <Button type="link" danger @click="onDelete(row)">
          {{ $t('common.actions.delete') }}
        </Button>
      </template>
    </Grid>
    <Modal
      v-model:open="modalVisible"
      :title="isEdit ? '编辑订单' : '新增订单'"
      :confirm-loading="submitting"
      @ok="onSubmit"
    >
      <Form
        ref="formRef"
        :model="form"
        :rules="rules"
        :label-col="{ span: 5 }"
        :wrapper-col="{ span: 18 }"
      >
        <FormItem label="name" name="name">
          <Input v-model:value="form.name" allow-clear placeholder="name" />
        </FormItem>
        <FormItem label="sn" name="sn">
          <Input v-model:value="form.sn" allow-clear placeholder="sn" />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
