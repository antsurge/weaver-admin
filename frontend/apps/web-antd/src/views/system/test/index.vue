<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { test } from '#/api/system/test';

import { reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, Form, FormItem, message, Modal, Select } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  createtestApi,
  deletetestApi,
  listtestApi,
  updatetestApi,
} from '#/api/system/test';
import { $t } from '#/locales';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

defineOptions({ name: 'TestManagement' });

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: 'Select',
        componentProps: {
          placeholder: '状态',
        },
        fieldName: 'status',
        label: '状态',
      },
    ],
  },
  gridOptions: {
    columns: [
      {
        field: 'status',
        align: 'center',
        title: '状态',
        minWidth: 150,
      },
      {
        field: 'createdAt',
        align: 'center',
        title: '创建时间',
        minWidth: 150,
        formatter: 'formatDateTime',
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
          const res = await listtestApi({
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

const formRef = ref();
const modalVisible = ref(false);
const submitting = ref(false);
const isEdit = ref(false);
const form = reactive<Partial<test>>({
  status: undefined,
});

const rules: Record<string, any> = {};

function onRefresh() {
  gridApi.query();
}

function onCreate() {
  isEdit.value = false;
  Object.keys(form).forEach((key) => {
    form[key as keyof typeof form] = undefined;
  });
  modalVisible.value = true;
}

function onEdit(row: test) {
  isEdit.value = true;
  Object.assign(form, row);
  modalVisible.value = true;
}

function onDelete(row: test) {
  Modal.confirm({
    title: $t('common.confirm'),
    content: $t('common.actions.delete'),
    async onOk() {
      await deletetestApi([row.id]);
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
      ? updatetestApi(form.id as string, form)
      : createtestApi(form));
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
    <Grid>
      <template #toolbar-actions>
        <Button type="primary" @click="onCreate">
          <Plus class="mr-1" />
          <span v-text="$t('common.actions.create')"></span>
        </Button>
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
      :title="isEdit ? '编辑测试2' : '新增测试2'"
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
        <FormItem label="状态" name="status">
          <Select v-model:value="form.status" allow-clear placeholder="状态" />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
