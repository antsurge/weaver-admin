<script lang="ts" setup>
import type { Recordable } from '@vben/types';

import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { PermissionMenuApi } from '#/api/permission/menu';

import { nextTick, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { IconifyIcon, Plus } from '@vben/icons';
import { $t } from '@vben/locales';

import { MenuBadge } from '@vben-core/menu-ui';

import { Button, message, Modal } from 'ant-design-vue';

import { PermissionAuthCode } from '#/access';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteMenuApi,
  getMenuTreeApi,
  updateMenuStatusApi,
} from '#/api/permission/menu';

import {
  PermissionTypeOptionsValueAction,
  useColumns,
  useFormOptions,
} from './data';
import Form from './modules/form/index.vue';

const [FormModel, formModelApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

const selectedRows = ref<PermissionMenuApi.PermissionMenu[]>([]);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: useFormOptions(),
  gridOptions: {
    columns: useColumns(onActionClick, onStatusChange),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: false,
    },
    proxyConfig: {
      autoLoad: true,
      ajax: {
        query: async (_params, formValues) => {
          const res = await getMenuTreeApi(formValues);
          const items = res?.items ?? [];
          // tree-config.transform=true 要求数据源为扁平结构，vxe 内部按 parentID 组装树
          return flattenMenuTree(items);
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
    treeConfig: {
      parentField: 'parentID',
      rowField: 'id',
      childrenField: 'children',
      mapChildrenField: '_children',
      transform: true,
    },
  } as VxeTableGridOptions,
  gridEvents: {
    checkboxChange() {
      nextTick(() => {
        onCheckboxChange();
      });
    },
    checkboxAll() {
      nextTick(() => {
        onCheckboxChange();
      });
    },
  },
});

function onCheckboxChange() {
  const checkboxRecords = gridApi.grid?.getCheckboxRecords?.() ?? [];
  selectedRows.value = checkboxRecords;
}

function onActionClick({
  code,
  row,
}: OnActionClickParams<PermissionMenuApi.PermissionMenu>) {
  switch (code) {
    case 'append': {
      onAppend(row);
      break;
    }
    case 'delete': {
      onDelete(row);
      break;
    }
    case 'edit': {
      onEdit(row);
      break;
    }
    default: {
      break;
    }
  }
}

function onRefresh() {
  gridApi.query();
}

function onEdit(row: PermissionMenuApi.PermissionMenu) {
  formModelApi.setData(row).open();
}

function onCreate() {
  formModelApi.setData({}).open();
}
function onAppend(row: PermissionMenuApi.PermissionMenu) {
  formModelApi.setData({ parentID: row.id }).open();
}

/**
 * 将树形菜单数据扁平化为数组（保留每行的 parentID 层级关系），
 * 供 tree-config.transform=true 使用，vxe 内部按 parentID 组装树。
 */
function flattenMenuTree(
  nodes: PermissionMenuApi.PermissionMenu[],
): PermissionMenuApi.PermissionMenu[] {
  const result: PermissionMenuApi.PermissionMenu[] = [];
  nodes.forEach((node) => {
    const { children, ...rest } = node;
    result.push(rest);
    if (children?.length) {
      result.push(...flattenMenuTree(children));
    }
  });
  return result;
}

async function onStatusChange(
  newStatus: string,
  row: PermissionMenuApi.PermissionMenu,
): Promise<boolean> {
  const status: Recordable<string> = {
    disabled: $t('common.disabled'),
    enabled: $t('common.enabled'),
  };

  return new Promise<boolean>((resolve) => {
    Modal.confirm({
      title: $t('permission.menu.actionMessage.switchStatus'),
      content: $t('permission.menu.actionMessage.switchStatusConfirm', [
        row.name,
        status[newStatus],
      ]),
      okText: $t('ui.actionTitle.confirm'),
      cancelText: $t('ui.actionTitle.cancel'),
      async onOk() {
        try {
          await updateMenuStatusApi(row.id, newStatus);
          resolve(true);
        } catch {
          message.error($t('ui.actionMessage.operationFailed'));
          resolve(false);
        }
      },
      onCancel() {
        resolve(false);
      },
    });
  });
}

function onDelete(row: PermissionMenuApi.PermissionMenu) {
  const hideLoading = message.loading({
    content: $t('ui.actionMessage.deleting', [row.name]),
    duration: 0,
    key: 'action_process_msg',
  });

  deleteMenuApi([row.id])
    .then(() => {
      message.success({
        content: $t('ui.actionMessage.deleteSuccess', [row.name]),
        key: 'action_process_msg',
      });
      onRefresh();
    })
    .catch(() => {
      message.error({
        content: $t('ui.actionMessage.operationFailed'),
        key: 'action_process_msg',
      });
      hideLoading();
    });
}

// 展开全部
function onExpandAll() {
  gridApi.grid?.setAllTreeExpand(true);
}

// 折叠全部
function onCollapseAll() {
  gridApi.grid?.clearTreeExpand();
}

// 批量删除
function onBatchDelete() {
  const rows = selectedRows.value;
  if (!rows?.length) return;
  const ids = rows.map((r) => r.id);
  const allNames = rows.map((r) => r.name ?? r.title ?? r.id);
  // 名称过多时截断展示，避免提示框过长
  const names =
    allNames.length > 3
      ? `${allNames.slice(0, 3).join('、')} 等 ${allNames.length} 项`
      : allNames.join('、');

  Modal.confirm({
    title: $t('ui.actionMessage.confirmDelete'),
    content: $t('ui.actionMessage.deleteConfirm', [names]),
    okText: $t('ui.actionTitle.confirm'),
    cancelText: $t('ui.actionTitle.cancel'),
    okType: 'danger',
    onOk() {
      const hideLoading = message.loading({
        content: $t('ui.actionMessage.deleting', [names]),
        duration: 0,
        key: 'batch_delete_msg',
      });
      return deleteMenuApi(ids)
        .then(() => {
          message.success({
            content: $t('ui.actionMessage.deleteSuccess', [names]),
            key: 'batch_delete_msg',
          });
          selectedRows.value = [];
          onRefresh();
        })
        .catch(() => {
          message.error({
            content: $t('ui.actionMessage.operationFailed'),
            key: 'batch_delete_msg',
          });
          hideLoading();
        });
    },
  });
}
</script>
<template>
  <Page auto-content-height>
    <FormModel @success="onRefresh" />
    <Grid>
      <template #toolbar-tools>
        <div class="flex gap-2">
          <Button
            v-access:code="[PermissionAuthCode.Menu.ExpandAll]"
            type="primary"
            class="inline-flex items-center"
            @click="onExpandAll"
          >
            <IconifyIcon
              icon="ant-design:column-height-outlined"
              class="size-5"
            />
            {{ $t('permission.menu.actionTitle.expandAll') }}
          </Button>
          <Button
            v-access:code="[PermissionAuthCode.Menu.CollapseAll]"
            type="primary"
            class="inline-flex items-center"
            @click="onCollapseAll"
          >
            <IconifyIcon
              icon="ant-design:column-width-outlined"
              class="size-5"
            />
            {{ $t('permission.menu.actionTitle.collapseAll') }}
          </Button>
          <Button
            v-access:code="[PermissionAuthCode.Menu.BatchDelete]"
            type="primary"
            class="inline-flex items-center"
            danger
            @click="onBatchDelete"
            :disabled="selectedRows.length === 0"
          >
            <IconifyIcon icon="ant-design:delete-outlined" class="size-5" />
            {{ $t('ui.actionTitle.delete') }}
          </Button>
          <Button
            v-access:code="[PermissionAuthCode.Menu.Create]"
            type="primary"
            class="inline-flex items-center"
            @click="onCreate"
          >
            <Plus class="size-5" />
            {{ $t('ui.actionTitle.create') }}
          </Button>
        </div>
      </template>
      <template #title="{ row }">
        <div class="menu-title-cell">
          <!-- 左侧 icon + title -->
          <div class="flex items-center gap-2">
            <div class="flex size-5 shrink-0 items-center justify-center">
              <IconifyIcon
                v-if="row.type === PermissionTypeOptionsValueAction"
                icon="carbon:security"
                class="size-4 shrink-0"
              />
              <IconifyIcon
                v-else-if="row?.icon"
                :icon="row.icon || 'carbon:circle-dash'"
                class="size-4 shrink-0"
              />
            </div>
            <span class="menu-title-text">
              {{ $t(row?.title) }}
            </span>
          </div>
          <!-- 右侧 Badge -->
          <MenuBadge
            v-if="row?.badgeType"
            class="menu-badge"
            :badge="row.badge"
            :badge-type="row.badgeType"
            :badge-variants="row.badgeVariants"
          />
        </div>
      </template>
    </Grid>
  </Page>
</template>
<style lang="scss" scoped>
.menu-title-cell {
  position: relative;
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: space-between;
  min-width: 0;
  height: 100%;

  .menu-title-text {
    display: inline-flex;
    align-items: center;
  }
}

.menu-badge {
  top: 50%;
  right: 0;
  transform: translateY(-50%);

  & > :deep(div) {
    padding-top: 0;
    padding-bottom: 0;
  }
}
</style>
