<script lang="ts" setup>
import type { MonitorApi } from '#/api/monitor/monitor';

import { onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Card,
  Col,
  Descriptions,
  message,
  Row,
  Spin,
  Statistic,
} from 'ant-design-vue';

import { getServerInfoApi } from '#/api/monitor/monitor';
import { $t } from '#/locales';

import { formatBytes, formatDuration } from '../use-format';

const loading = ref(false);
const info = ref<MonitorApi.ServerInfo>();

async function load() {
  loading.value = true;
  try {
    info.value = await getServerInfoApi();
  } catch (error) {
    console.error(error);
    message.error($t('monitor.message.loadFailed'));
  } finally {
    loading.value = false;
  }
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
              :title="$t('monitor.server.uptime')"
              :value="formatDuration(info?.uptimeSeconds)"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.server.numGoroutine')"
              :value="info?.numGoroutine ?? 0"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.server.heapAlloc')"
              :value="formatBytes(info?.heapAlloc)"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card :bordered="false">
            <Statistic
              :title="$t('monitor.server.numGC')"
              :value="info?.numGC ?? 0"
            />
          </Card>
        </Col>
      </Row>

      <Card class="mt-4" :title="$t('monitor.server.title')" :bordered="false">
        <Descriptions bordered :column="2" size="small">
          <Descriptions.Item :label="$t('monitor.server.hostname')">
            {{ info?.hostname }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.startTime')">
            {{ info?.startTime }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.os')">
            {{ info?.os }} / {{ info?.arch }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.goVersion')">
            {{ info?.goVersion }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.numCPU')">
            {{ info?.numCPU }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.goMaxProcs')">
            {{ info?.goMaxProcs }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.heapSys')">
            {{ formatBytes(info?.heapSys) }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.heapInuse')">
            {{ formatBytes(info?.heapInuse) }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.stackInuse')">
            {{ formatBytes(info?.stackInuse) }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.sys')">
            {{ formatBytes(info?.sys) }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.totalAlloc')">
            {{ formatBytes(info?.totalAlloc) }}
          </Descriptions.Item>
          <Descriptions.Item :label="$t('monitor.server.gcPause')">
            {{ info?.gcPauseTotalMs ?? 0 }} ms
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </Spin>
  </Page>
</template>
