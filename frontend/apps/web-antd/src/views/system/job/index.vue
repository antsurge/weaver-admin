<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemJobApi } from '#/api/system/job';

import { nextTick, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal, Tag } from 'ant-design-vue';

import { SystemAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteJobApi,
  getJobListApi,
  runJobApi,
  updateJobStatusApi,
} from '#/api/system/job';
import { $t } from '#/locales';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import { useFormOptions, useJobColumns } from './data';
import JobForm from './modules/form/index.vue';

const [JobFormModal, jobFormApi] = useVbenModal({
  connectedComponent: JobForm,
  destroyOnClose: true,
});

const selectedRows = ref<SystemJobApi.Job[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useJobColumns(onActionClick, onStatusChange),
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
          const res = await getJobListApi({
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
  jobFormApi.setData({}).open();
}

function onActionClick({ code, row }: { code: string; row: SystemJobApi.Job }) {
  switch (code) {
    case 'delete': {
      onDelete([row]);
      break;
    }
    case 'edit': {
      jobFormApi.setData(row).open();
      break;
    }
    case 'run': {
      onRun(row);
      break;
    }
  }
}

function onStatusChange(
  newStatus: SystemJobApi.Job['status'],
  row: SystemJobApi.Job,
): Promise<boolean | undefined> {
  const statusText: Record<string, string> = {
    enabled: $t('common.enabled'),
    disabled: $t('common.disabled'),
  };

  return new Promise<boolean | undefined>((resolve) => {
    Modal.confirm({
      title: $t('common.confirm'),
      content: `${$t('system.job.message.statusConfirm')}【${statusText[newStatus]}】吗？`,
      async onOk() {
        try {
          await updateJobStatusApi(row.id, newStatus);
          resolve(true);
        } catch {
          resolve(false);
        }
      },
      onCancel() {
        resolve(false);
      },
    });
  });
}

function onRun(row: SystemJobApi.Job) {
  Modal.confirm({
    title: $t('system.job.confirm.runTitle'),
    content: $t('system.job.confirm.runContent', [row.name]),
    async onOk() {
      await runJobApi(row.id);
      message.success($t('system.job.message.runSuccess'));
    },
  });
}

function onDelete(rows: SystemJobApi.Job[]) {
  Modal.confirm({
    title: $t('system.job.confirm.deleteTitle'),
    content: $t('system.job.confirm.deleteContent', [rows.length]),
    async onOk() {
      const ids = rows.map((item) => item.id);
      await deleteJobApi(ids);
      message.success($t('system.job.message.deleteSuccess'));
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
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            v-access:code="[SystemAuthCode.Job.Create]"
            type="primary"
            @click="onCreate"
          >
            <Plus class="mr-1" />
            {{ $t('common.actions.create') }}
          </Button>
          <Button
            v-access:code="[SystemAuthCode.Job.BatchDelete]"
            danger
            :disabled="selectedRows.length === 0"
            @click="onBatchDelete"
          >
            {{ $t('system.job.operation.batchDelete') }}
          </Button>
        </div>
      </template>
      <template #job_type="{ row }">
        <Tag :color="row.jobType === 'http' ? 'blue' : 'purple'">
          {{ row.jobType?.toUpperCase() }}
        </Tag>
      </template>
      <template #misfire_policy="{ row }">
        <Tag>
          {{
            $t(
              `system.job.misfire_options.${row.misfirePolicy || 'immediately'}`,
            )
          }}
        </Tag>
      </template>
    </Grid>
    <JobFormModal @success="onRefresh" />
  </Page>
</template>
