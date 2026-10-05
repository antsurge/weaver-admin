<script lang="ts" setup>
import type { MonitorApi } from '#/api/monitor/monitor';

import { onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Alert,
  Button,
  Card,
  Col,
  Input,
  message,
  Row,
  Select,
  Spin,
  Statistic,
  Table,
  Tag,
} from 'ant-design-vue';

import { getApiMetricsApi } from '#/api/monitor/monitor';
import { $t } from '#/locales';

import { formatDuration } from '../use-format';

const HTTP_METHODS = [
  'GET',
  'POST',
  'PUT',
  'DELETE',
  'PATCH',
  'OPTIONS',
  'HEAD',
];

const loading = ref(false);
const result = ref<MonitorApi.ApiMetricsResult>();
const query = reactive<{ method?: string; path: string }>({
  method: undefined,
  path: '',
});

const columns = [
  {
    align: 'center',
    dataIndex: 'method',
    key: 'method',
    title: $t('monitor.api.method'),
    width: 90,
  },
  { dataIndex: 'path', key: 'path', title: $t('monitor.api.path') },
  {
    align: 'center',
    dataIndex: 'total',
    key: 'total',
    title: $t('monitor.api.total'),
    width: 100,
  },
  {
    align: 'center',
    dataIndex: 'success',
    key: 'success',
    title: $t('monitor.api.success'),
    width: 100,
  },
  {
    align: 'center',
    dataIndex: 'error',
    key: 'error',
    title: $t('monitor.api.error'),
    width: 90,
  },
  {
    align: 'center',
    dataIndex: 'errorRate',
    key: 'errorRate',
    title: $t('monitor.api.errorRate'),
    width: 110,
  },
  {
    align: 'center',
    dataIndex: 'avgMs',
    key: 'avgMs',
    title: $t('monitor.api.avgMs'),
    width: 110,
  },
  {
    align: 'center',
    dataIndex: 'maxMs',
    key: 'maxMs',
    title: $t('monitor.api.maxMs'),
    width: 110,
  },
];

function rowKey(record: MonitorApi.ApiMetric): string {
  return `${record.method} ${record.path}`;
}

async function load() {
  loading.value = true;
  try {
    result.value = await getApiMetricsApi({
      method: query.method,
      path: query.path,
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
  query.method = undefined;
  query.path = '';
  load();
}

onMounted(load);
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Row :gutter="16">
        <Col :span="8">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.api.totalRequests')"
              :value="result?.totalRequests ?? 0"
            />
          </Card>
        </Col>
        <Col :span="8">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.api.totalErrors')"
              :value="result?.totalErrors ?? 0"
            />
          </Card>
        </Col>
        <Col :span="8">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.api.uptime')"
              :value="formatDuration(result?.uptimeSeconds)"
            />
          </Card>
        </Col>
      </Row>

      <Card class="mt-4" :title="$t('monitor.api.title')" :bordered="false">
        <Alert
          class="mb-3"
          :message="$t('monitor.api.tip')"
          show-icon
          type="info"
        />
        <div class="mb-3 flex items-center gap-2">
          <Select
            v-model:value="query.method"
            :options="HTTP_METHODS.map((m) => ({ label: m, value: m }))"
            :placeholder="$t('monitor.api.methodPlaceholder')"
            allow-clear
            class="w-40"
          />
          <Input
            v-model:value="query.path"
            :placeholder="$t('monitor.api.pathPlaceholder')"
            allow-clear
            class="max-w-[360px]"
            @press-enter="onSearch"
          />
          <Button type="primary" @click="onSearch">
            {{ $t('common.actions.search') }}
          </Button>
          <Button @click="onReset">{{ $t('common.actions.reset') }}</Button>
        </div>
        <Table
          :columns="columns"
          :data-source="result?.items ?? []"
          :pagination="{ pageSize: 20, showSizeChanger: true }"
          :row-key="rowKey"
          :scroll="{ x: 1100 }"
          class="api-table"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'method'">
              <Tag :color="record.method === 'GET' ? 'blue' : 'geekblue'">
                {{ record.method }}
              </Tag>
            </template>
            <template v-else-if="column.key === 'error'">
              <span :class="record.error > 0 ? 'text-red-500' : ''">
                {{ record.error }}
              </span>
            </template>
            <template v-else-if="column.key === 'errorRate'">
              {{ record.errorRate }}%
            </template>
            <template v-else-if="column.key === 'avgMs'">
              {{ record.avgMs }}
            </template>
            <template v-else-if="column.key === 'maxMs'">
              {{ record.maxMs }}
            </template>
          </template>
        </Table>
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
.api-table :deep(.ant-table-thead > tr > th) {
  text-align: center;
}
</style>
