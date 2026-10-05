<script lang="ts" setup>
import type { MonitorApi } from '#/api/monitor/monitor';

import { onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Descriptions,
  Input,
  message,
  Spin,
  Table,
} from 'ant-design-vue';

import { getDatabaseInfoApi } from '#/api/monitor/monitor';
import { $t } from '#/locales';

import { formatBytes } from '../use-format';

const loading = ref(false);
const info = ref<MonitorApi.DatabaseInfo>();
const keyword = ref('');
const filteredTables = ref<MonitorApi.DbTableStat[]>([]);

const columns = [
  {
    align: 'center',
    dataIndex: 'name',
    key: 'name',
    title: $t('monitor.database.tableName'),
  },
  {
    align: 'center',
    dataIndex: 'rows',
    key: 'rows',
    title: $t('monitor.database.rows'),
    width: 140,
  },
  {
    align: 'center',
    dataIndex: 'dataSize',
    key: 'dataSize',
    title: $t('monitor.database.dataSize'),
    width: 140,
  },
  {
    align: 'center',
    dataIndex: 'indexSize',
    key: 'indexSize',
    title: $t('monitor.database.indexSize'),
    width: 140,
  },
];

function filterTables() {
  const tables = info.value?.tables ?? [];
  const kw = keyword.value.trim().toLowerCase();
  filteredTables.value = kw
    ? tables.filter((table) => table.name.toLowerCase().includes(kw))
    : tables;
}

async function load() {
  loading.value = true;
  try {
    info.value = await getDatabaseInfoApi();
    filterTables();
  } catch (error) {
    console.error(error);
    message.error($t('monitor.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  filterTables();
}

function onReset() {
  keyword.value = '';
  filterTables();
}

onMounted(load);
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Card :title="$t('monitor.database.title')" :bordered="false">
        <Descriptions bordered :column="3" size="small">
          <Descriptions.Item :label="$t('monitor.database.driver')">
            {{ info?.driver }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.database.version')">
            {{ info?.version }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.database.database')">
            {{ info?.database }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.database.tableCount')">
            {{ info?.tableCount }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.database.openConnections')">
            {{ info?.openConnections }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.database.inUse')">
            {{ info?.inUse }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.database.idle')">
            {{ info?.idle }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.database.maxOpenConnections')">
            {{
              info?.maxOpenConnections === 0 ? '∞' : info?.maxOpenConnections
            }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.database.waitCount')">
            {{ info?.waitCount }} / {{ info?.waitDurationMs }} ms
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </Spin>

    <Card class="mt-4" :bordered="false">
      <div class="mb-3 flex items-center gap-2">
        <Input
          v-model:value="keyword"
          :placeholder="$t('monitor.database.tableNamePlaceholder')"
          allow-clear
          class="max-w-[360px]"
          @press-enter="onSearch"
        />
        <Button type="primary" @click="onSearch">
          {{ $t('common.actions.search') }}
        </Button>
        <Button @click="onReset">
          {{ $t('monitor.database.reset') }}
        </Button>
      </div>
      <Table
        :columns="columns"
        :data-source="filteredTables"
        :pagination="{ pageSize: 20, showSizeChanger: true }"
        row-key="name"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'dataSize'">
            {{ formatBytes(record.dataSize) }}
          </template>
          <template v-else-if="column.key === 'indexSize'">
            {{ formatBytes(record.indexSize) }}
          </template>
        </template>
      </Table>
    </Card>
  </Page>
</template>
