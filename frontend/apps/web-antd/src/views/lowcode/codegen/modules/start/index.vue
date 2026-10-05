<script lang="ts" setup>
import { IconifyIcon } from '@vben/icons';

import { $t } from '#/locales';

defineEmits<{
  (e: 'create'): void;
  (e: 'table'): void;
  (e: 'records'): void;
}>();
</script>

<template>
  <div class="mx-auto max-w-[980px] py-10">
    <!-- 标题 -->
    <div class="mb-8 text-center">
      <h1 class="mb-2 text-3xl font-semibold text-gray-800 dark:text-gray-100">
        {{ $t('lowcode.codegen.start.title') }}
      </h1>
      <p class="text-gray-500 dark:text-gray-400">
        {{ $t('lowcode.codegen.start.subtitle') }}
      </p>
    </div>

    <!-- 三个入口卡片 -->
    <div class="grid grid-cols-1 gap-5 sm:grid-cols-3">
      <div
        class="group cursor-pointer rounded-lg border border-blue-100 bg-blue-50/70 p-6 transition-all hover:-translate-y-1 hover:border-blue-300 hover:shadow-lg dark:border-blue-900/60 dark:bg-blue-950/30"
        @click="$emit('create')"
      >
        <div
          class="mb-3 flex h-12 w-12 items-center justify-center rounded-md bg-blue-500 text-white shadow-md transition-transform group-hover:scale-110"
        >
          <IconifyIcon class="text-xl" icon="ant-design:plus-outlined" />
        </div>
        <h3 class="mb-1 text-lg font-semibold text-gray-800 dark:text-gray-100">
          {{ $t('lowcode.codegen.start.create.title') }}
        </h3>
        <p class="text-sm leading-relaxed text-gray-500 dark:text-gray-400">
          {{ $t('lowcode.codegen.start.create.desc') }}
        </p>
      </div>

      <div
        class="group cursor-pointer rounded-lg border border-blue-100 bg-blue-50/70 p-6 transition-all hover:-translate-y-1 hover:border-blue-300 hover:shadow-lg dark:border-blue-900/60 dark:bg-blue-950/30"
        @click="$emit('table')"
      >
        <div
          class="mb-3 flex h-12 w-12 items-center justify-center rounded-md bg-blue-500 text-white shadow-md transition-transform group-hover:scale-110"
        >
          <IconifyIcon class="text-xl" icon="ant-design:database-outlined" />
        </div>
        <h3 class="mb-1 text-lg font-semibold text-gray-800 dark:text-gray-100">
          {{ $t('lowcode.codegen.start.table.title') }}
        </h3>
        <p class="text-sm leading-relaxed text-gray-500 dark:text-gray-400">
          {{ $t('lowcode.codegen.start.table.desc') }}
        </p>
      </div>

      <div
        class="group cursor-pointer rounded-lg border border-blue-100 bg-blue-50/70 p-6 transition-all hover:-translate-y-1 hover:border-blue-300 hover:shadow-lg dark:border-blue-900/60 dark:bg-blue-950/30"
        @click="$emit('records')"
      >
        <div
          class="mb-3 flex h-12 w-12 items-center justify-center rounded-md bg-blue-500 text-white shadow-md transition-transform group-hover:scale-110"
        >
          <IconifyIcon class="text-xl" icon="ant-design:file-text-outlined" />
        </div>
        <h3 class="mb-1 text-lg font-semibold text-gray-800 dark:text-gray-100">
          {{ $t('lowcode.codegen.start.records.title') }}
        </h3>
        <p class="text-sm leading-relaxed text-gray-500 dark:text-gray-400">
          {{ $t('lowcode.codegen.start.records.desc') }}
        </p>
      </div>
    </div>

    <!-- 快速体验区 -->
    <div
      class="mt-10 rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-800"
    >
      <h3 class="mb-4 text-base font-semibold text-gray-700 dark:text-gray-200">
        {{ $t('lowcode.codegen.start.quick.title') }}
      </h3>
      <ol class="space-y-3 text-sm text-gray-600 dark:text-gray-300">
        <li v-for="(_, idx) in 5" :key="idx" class="flex gap-2">
          <span
            class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-blue-500 text-xs text-white"
            >{{ idx + 1 }}</span
          >
          <span>{{ $t(`lowcode.codegen.start.quick.step${idx + 1}`) }}</span>
        </li>
      </ol>
      <div
        class="mt-4 rounded-md border border-orange-300 bg-orange-50 px-4 py-3 text-sm text-orange-600 dark:border-orange-700/60 dark:bg-orange-900/20 dark:text-orange-400"
      >
        {{ $t('lowcode.codegen.start.quick.warning') }}
      </div>
    </div>
  </div>
</template>
