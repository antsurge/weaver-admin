<script lang="ts" setup>
import type { MonitorApi } from '#/api/monitor/monitor';

import { onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Col,
  Input,
  message,
  Row,
  Spin,
  Statistic,
  Table,
  Tag,
} from 'ant-design-vue';

import { getJobMonitorApi } from '#/api/monitor/monitor';
import { $t } from '#/locales';

const loading = ref(false);
const info = ref<MonitorApi.JobMonitorInfo>();
const query = reactive<{ group: string; name: string }>({
  name: '',
  group: '',
});

const columns = [
  {
    dataIndex: 'jobName',
    key: 'jobName',
    title: $t('monitor.job.jobName'),
    align: 'center',
    width: 200,
  },
  {
    dataIndex: 'jobGroup',
    key: 'jobGroup',
    title: $t('monitor.job.jobGroup'),
    width: 140,
    align: 'center',
  },
  {
    dataIndex: 'status',
    key: 'status',
    title: $t('monitor.job.status'),
    width: 110,
    align: 'center',
  },
  {
    dataIndex: 'startTime',
    key: 'startTime',
    title: $t('monitor.job.startTime'),
    width: 200,
    align: 'center',
  },
  {
    dataIndex: 'duration',
    key: 'duration',
    title: $t('monitor.job.duration'),
    width: 110,
    align: 'center',
  },
  { dataIndex: 'message', key: 'message', title: $t('monitor.job.message') },
];

async function load() {
  loading.value = true;
  try {
    info.value = await getJobMonitorApi({
      name: query.name || undefined,
      group: query.group || undefined,
    });
  } catch (error) {
    console.error(error);
    message.error($t('monitor.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  load();
}

function onReset() {
  query.name = '';
  query.group = '';
  load();
}

onMounted(load);
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Row :gutter="16">
        <Col :span="6">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.job.totalJobs')"
              :value="info?.totalJobs ?? 0"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.job.enabledJobs')"
              :value="info?.enabledJobs ?? 0"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.job.runningJobs')"
              :value="info?.runningJobs ?? 0"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.job.successRate')"
              :value="info?.successRate ?? 0"
              suffix="%"
            />
          </Card>
        </Col>
      </Row>

      <Card class="mt-4" :bordered="false">
        <Row :gutter="16">
          <Col :span="8">
            <Statistic
              :title="$t('monitor.job.totalRuns')"
              :value="info?.totalRuns ?? 0"
            />
          </Col>
          <Col :span="8">
            <Statistic
              :title="$t('monitor.job.successRuns')"
              :value="info?.successRuns ?? 0"
              :value-style="{ color: '#3f8600' }"
            />
          </Col>
          <Col :span="8">
            <Statistic
              :title="$t('monitor.job.failedRuns')"
              :value="info?.failedRuns ?? 0"
              :value-style="{ color: '#cf1322' }"
            />
          </Col>
        </Row>
      </Card>

      <Card
        class="mt-4"
        :title="$t('monitor.job.recentRuns')"
        :bordered="false"
      >
        <div class="mb-3 flex items-center gap-2">
          <Input
            v-model:value="query.name"
            :placeholder="$t('monitor.job.namePlaceholder')"
            allow-clear
            class="w-48"
            @press-enter="onSearch"
          />
          <Input
            v-model:value="query.group"
            :placeholder="$t('monitor.job.groupPlaceholder')"
            allow-clear
            class="w-48"
            @press-enter="onSearch"
          />
          <Button type="primary" @click="onSearch">
            {{ $t('common.actions.search') }}
          </Button>
          <Button @click="onReset">{{ $t('common.actions.reset') }}</Button>
        </div>
        <Table
          class="job-table"
          :columns="columns"
          :data-source="info?.recentRuns ?? []"
          :pagination="false"
          :scroll="{ x: 1100 }"
          row-key="id"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <Tag :color="record.status === 'success' ? 'green' : 'red'">
                {{ record.status }}
              </Tag>
            </template>
            <template v-else-if="column.key === 'duration'">
              {{ record.duration }} ms
            </template>
          </template>
        </Table>
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
.job-table :deep(.ant-table-thead > tr > th) {
  text-align: center;
}
</style>
