<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { TreeDataItem } from 'ant-design-vue/es/tree/Tree';

import { computed, onMounted, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';
import { message } from 'ant-design-vue';

import { Tree } from 'ant-design-vue';
import {
  getRoleApi,
  createRoleApi,
  updateRoleApi,
  bindMenusForRoleApi,
} from '#/api/permission/role';
import { getMenuTreeApi } from '#/api/permission/menu';
import type { PermissionRoleApi } from '#/api/permission/role';
import { useVbenForm } from '#/adapter/form';

import { $t } from '#/locales';

import { nameRule, codeRule } from './rules';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<PermissionRoleApi.Role>();

// 菜单树相关状态
const menuTreeData = ref<TreeDataItem[]>([]);
const checkedKeys = ref<string[]>([]);
const halfCheckedKeys = ref<string[]>([]);
const menuTreeLoading = ref(false);

// 菜单权限校验状态
const menuValidateStatus = ref<'' | 'error'>('');
const menuHelpText = ref('');

// 加载菜单树
async function loadMenuTree() {
  menuTreeLoading.value = true;
  try {
    const res = await getMenuTreeApi({});
    // 转换为 Tree 组件需要的格式
    menuTreeData.value = transformToTreeData(res.items || []);
  } catch (error) {
    console.error('加载菜单树失败:', error);
  } finally {
    menuTreeLoading.value = false;
  }
}

// 转换后端数据为 Tree 组件格式
function transformToTreeData(items: any[]): TreeDataItem[] {
  return items.map(item => ({
    key: item.id,
    title: item.name,
    children: item.children ? transformToTreeData(item.children) : undefined,
  }));
}

// 树节点选中变化
function onTreeCheck(keys: string[], info: any) {
  checkedKeys.value = keys;
  halfCheckedKeys.value = info.halfCheckedKeys;

  // 当用户选择了菜单时，清除错误状态
  if (keys.length > 0) {
    menuValidateStatus.value = '';
    menuHelpText.value = '';
  }
}

// 校验菜单选择是否为空
function validateMenuSelection(): boolean {
  if (checkedKeys.value.length === 0) {
    menuValidateStatus.value = 'error';
    menuHelpText.value = $t('permission.role.rules.menuRequired') || '请至少选择一个菜单权限';
    return false;
  }
  menuValidateStatus.value = '';
  menuHelpText.value = '';
  return true;
}

const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('permission.role.fields.name'),
    component: 'Input',
    rules: nameRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('permission.role.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'code',
    label: $t('permission.role.fields.code'),
    component: 'Input',
    rules: codeRule,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('permission.role.fields.code'),
      ]),
    },
  },
  {
    fieldName: 'weight',
    label: $t('permission.role.fields.weight'),
    component: 'InputNumber',
    defaultValue: 0,
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('permission.role.fields.weight'),
      ]),
      min: 0,
      step: 1,
      precision: 0,
      stringMode: false,
      style: { width: '100%' },
    },
  },
  {
    fieldName: 'status',
    label: $t('permission.role.fields.status'),
    component: 'Switch',
    defaultValue: 'enabled',
    componentProps: {
      class: 'w-auto',
      checkedChildren: $t('common.enabled'),
      checkedValue: 'enabled',
      unCheckedChildren: $t('common.disabled'),
      unCheckedValue: 'disabled',
    },
  },
  {
    fieldName: 'remark',
    label: $t('permission.role.fields.remark'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('permission.role.form_placeholder.remark'),
    },
  },
];

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    colon: true,
    formItemClass: 'col-span-2 md:col-span-2',
    labelWidth: 90,
  },
  schema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2 gap-x-4',
});

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange:async (isOpen) => {
    if (!isOpen) return;

    // 重置菜单校验状态
    menuValidateStatus.value = '';
    menuHelpText.value = '';

    // 加载菜单树（只在首次打开时加载）
    if (menuTreeData.value.length === 0) {
      await loadMenuTree();
    }

    const data = modalApi.getData<PermissionRoleApi.Role>();
    // 编辑
    if (data?.id) {
      modalApi.lock();
      try {
        const res = await getRoleApi(data.id);
        formData.value = res;
        formApi.setValues(res);
        // 回显已绑定的菜单
        checkedKeys.value = res.menuIds || [];
      } finally {
        modalApi.unlock();
      }
    } else {
      // 👉 新增
      formData.value = undefined;
      formApi.resetForm();
      // 清空菜单选择
      checkedKeys.value = [];
      halfCheckedKeys.value = [];
    }
  },
});

async function onSubmit() {
  // 1. 校验表单
  const { valid } = await formApi.validate();
  if (!valid) return;

  // 2. 校验菜单权限（必填）
  if (!validateMenuSelection()) {
    message.warning($t('permission.role.rules.menuRequired') || '请至少选择一个菜单权限');
    return;
  }

  modalApi.lock();
  const data = await formApi.getValues<PermissionRoleApi.Role>();
  try {
    if (formData.value?.id) {
      // 更新角色基本信息
      await updateRoleApi(formData.value.id, data);
      // 重新绑定菜单（全量替换）
      await bindMenusForRoleApi(formData.value.id, checkedKeys.value);
    } else {
      // 创建角色
      const res = await createRoleApi(data);
      // 绑定菜单到新创建的角色
      if (res?.id) {
        await bindMenusForRoleApi(res.id, checkedKeys.value);
      }
    }
    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('permission.role.name')])
    : $t('ui.actionTitle.create', [$t('permission.role.name')]),
);

// 组件挂载时加载菜单树
onMounted(() => {
  loadMenuTree();
});
</script>

<template>
  <Modal class="w-full max-w-[800px]" :title="getModalTitle">
    <div class="mx-4">
      <!-- 基本信息表单 -->
      <Form class="mb-6" :layout="isHorizontal ? 'horizontal' : 'vertical'" />

      <!-- 菜单权限选择（必填） -->
      <div class="border-t pt-4">
        <div class="text-base font-medium mb-3 text-gray-700 flex items-center gap-1">
          {{ $t('permission.role.fields.menuPermission') || '菜单权限' }}
          <span class="text-red-500">*</span>
        </div>
        <div v-if="menuTreeLoading" class="flex justify-center py-8">
          <a-spin tip="加载菜单树..." />
        </div>
        <div
          v-else-if="menuTreeData.length > 0"
          class="max-h-[400px] overflow-auto rounded p-4"
          :class="menuValidateStatus === 'error' ? 'border-red-500 border bg-red-50' : 'border border-gray-200'"
        >
          <Tree
            v-model:checkedKeys="checkedKeys"
            :tree-data="menuTreeData"
            :checkable="true"
            :default-expand-all="true"
            :selectable="false"
            @check="onTreeCheck"
          >
            <template #title="{ title }">
              <span>{{ title }}</span>
            </template>
          </Tree>
          <!-- 错误提示 -->
          <div v-if="menuValidateStatus === 'error'" class="mt-2 text-red-500 text-sm">
            {{ menuHelpText }}
          </div>
        </div>
        <div v-else class="text-center py-8 text-gray-400">
          暂无菜单数据
        </div>
      </div>
    </div>
  </Modal>
</template>
