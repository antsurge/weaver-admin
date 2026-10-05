<script lang="ts" setup>
import type { TableColumnsType } from 'ant-design-vue';

import type { MonitorApi } from '#/api/monitor/monitor';

import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import { useClipboard } from '@vueuse/core';
import {
  Button,
  Card,
  Descriptions,
  Input,
  message,
  Spin,
  Table,
  Tooltip,
  TypographyText,
} from 'ant-design-vue';

import { getCacheInfoApi, getCacheKeysApi } from '#/api/monitor/monitor';
import { $t } from '#/locales';
import { DEFAULT_PAGE_SIZE } from '#/types/pagination';

import { formatBytes, formatDuration, formatTtl } from '../use-format';

const loading = ref(false);
const info = ref<MonitorApi.CacheInfo>();

const keyLoading = ref(false);
const keys = ref<MonitorApi.CacheKey[]>([]);
const total = ref(0);
const pattern = ref('');
const page = ref(1);
const pageSize = ref(DEFAULT_PAGE_SIZE);
const { copy } = useClipboard({ legacy: true });

const columns = computed<TableColumnsType>(() => [
  {
    align: 'center',
    dataIndex: 'key',
    key: 'key',
    title: $t('monitor.cache.key'),
    width: 280,
  },
  {
    align: 'center',
    dataIndex: 'type',
    key: 'type',
    title: $t('monitor.cache.type'),
    width: 160,
  },
  {
    align: 'center',
    dataIndex: 'value',
    key: 'value',
    title: $t('monitor.cache.value'),
    width: 260,
  },
  {
    align: 'center',
    dataIndex: 'ttlSeconds',
    key: 'ttl',
    title: $t('monitor.cache.ttl'),
    width: 260,
  },
  {
    align: 'center',
    dataIndex: 'size',
    key: 'size',
    title: $t('monitor.cache.size'),
    width: 220,
  },
]);

async function loadInfo() {
  loading.value = true;
  try {
    info.value = await getCacheInfoApi();
  } catch (error) {
    console.error(error);
    message.error($t('monitor.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

/**
 * 将用户输入转换为 Redis SCAN 匹配模式。
 * 输入为空或退出时，使用 "*"。若未包含通配符，则自动拼接为 "*keyword*" 实现模糊搜索。
 */
function toScanPattern(input: string): string {
  const keyword = input.trim();
  if (!keyword) {
    return '*';
  }
  if (keyword.includes('*') || keyword.includes('?')) {
    return keyword;
  }
  return `*${keyword}*`;
}

async function loadKeys() {
  keyLoading.value = true;
  try {
    const res = await getCacheKeysApi({
      currentPage: page.value,
      pageSize: pageSize.value,
      pattern: toScanPattern(pattern.value),
    });
    keys.value = res.items;
    total.value = res.total;
  } catch (error) {
    console.error(error);
    message.error($t('monitor.message.loadFailed'));
  } finally {
    keyLoading.value = false;
  }
}

function onSearch() {
  page.value = 1;
  loadKeys();
}

async function copyText(text: string | undefined) {
  if (!text) {
    return;
  }
  await copy(text);
  message.success($t('monitor.cache.copied'));
}

function onTableChange(pagination: any) {
  page.value = pagination.current;
  pageSize.value = pagination.pageSize;
  loadKeys();
}

onMounted(() => {
  loadInfo();
  loadKeys();
});
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Card :title="$t('monitor.cache.title')" :bordered="false">
        <Descriptions bordered :column="3" size="small">
          <Descriptions.Item :label="$t('monitor.cache.version')">
            {{ info?.version }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.cache.mode')">
            {{ info?.mode }} / {{ info?.role }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.cache.uptime')">
            {{ formatDuration(info?.uptimeSeconds) }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.cache.usedMemory')">
            {{ info?.usedMemoryHuman || formatBytes(info?.usedMemory) }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.cache.connectedClients')">
            {{ info?.connectedClients }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.cache.opsPerSec')">
            {{ info?.opsPerSec }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.cache.totalKeys')">
            {{ info?.totalKeys }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.cache.hitRate')">
            {{ info?.hitRate }}%
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.cache.totalCommands')">
            {{ info?.totalCommands }}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </Spin>

    <Card class="mt-4" :bordered="false">
      <div class="mb-3 flex items-center gap-2">
        <Input
          v-model:value="pattern"
          :placeholder="$t('monitor.cache.patternPlaceholder')"
          allow-clear
          class="max-w-[360px]"
          @press-enter="onSearch"
        />
        <Tooltip :title="$t('monitor.cache.fuzzyTip')">
          <Button type="link" size="small" class="p-0">?</Button>
        </Tooltip>
        <Button type="primary" @click="onSearch">
          {{ $t('common.actions.search') }}
        </Button>
        <Button @click="((pattern = ''), onSearch())">
          {{ $t('monitor.cache.allKeys') }}
        </Button>
      </div>
      <Table
        :columns="columns"
        :data-source="keys"
        :loading="keyLoading"
        :pagination="{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
        }"
        class="cache-table"
        table-layout="fixed"
        :scroll="{ x: 1180 }"
        row-key="rawKey"
        size="small"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'key'">
            <div class="flex w-full items-center gap-1">
              <Tooltip :title="record.key" placement="topLeft">
                <TypographyText class="cache-ellipsis min-w-0 flex-1">
                  {{ record.key }}
                </TypographyText>
              </Tooltip>
              <Tooltip :title="$t('monitor.cache.copyKey')">
                <Button
                  type="link"
                  size="small"
                  class="shrink-0 p-0"
                  @click="copyText(record.rawKey || record.key)"
                >
                  <IconifyIcon icon="lucide:copy" class="size-4" />
                </Button>
              </Tooltip>
            </div>
          </template>
          <template v-else-if="column.key === 'value'">
            <div class="flex w-full items-center gap-1">
              <Tooltip
                v-if="record.value"
                :title="record.value"
                placement="topLeft"
              >
                <TypographyText class="cache-ellipsis min-w-0 flex-1">
                  {{ record.value }}
                </TypographyText>
              </Tooltip>
              <span v-else class="min-w-0 flex-1 text-center">-</span>
              <Tooltip
                v-if="record.value"
                :title="$t('monitor.cache.copyValue')"
              >
                <Button
                  type="link"
                  size="small"
                  class="shrink-0 p-0"
                  @click="copyText(record.value)"
                >
                  <IconifyIcon icon="lucide:copy" class="size-4" />
                </Button>
              </Tooltip>
            </div>
          </template>
          <template v-else-if="column.key === 'ttl'">
            {{ formatTtl(record.ttlSeconds) }}
          </template>
          <template v-else-if="column.key === 'size'">
            {{ record.size < 0 ? '-' : formatBytes(record.size) }}
          </template>
        </template>
      </Table>
    </Card>
  </Page>
</template>

<style scoped>
.cache-table :deep(.ant-table-thead > tr > th) {
  text-align: center;
}

/* 长文本单行截断，避免溢出延伸 */
.cache-table :deep(.cache-ellipsis) {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
