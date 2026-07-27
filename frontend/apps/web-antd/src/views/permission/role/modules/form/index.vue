<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { TreeDataItem } from 'ant-design-vue/es/tree/Tree';
import type { Recordable } from '@vben/types';

import { computed, onMounted, ref } from 'vue';

import { Tree, useVbenModal } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';
import { Spin, message } from 'ant-design-vue';

import {
  getRoleApi,
  createRoleApi,
  updateRoleApi,
  bindMenusForRoleApi,
} from '#/api/permission/role';
import { PermissionTypeOptionsValueAction } from "#/views/permission/menu/data"
import { getMenuTreeApi } from '#/api/permission/menu';
import type { PermissionRoleApi } from '#/api/permission/role';
import { useVbenForm } from '#/adapter/form';

import { $t } from '#/locales';
import { $te } from '@vben/locales';

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

function getNodeClass(node: Recordable<any>) {
 const classes: string[] = [];
  if (node.value?.type === PermissionTypeOptionsValueAction) {
    classes.push('inline-flex');
  }

  return classes.join(' ');
}

// 转换后端数据为 Tree 组件格式
// 注意：treeData 项会被原样塞进 reka-ui InnerFlattenItem.value（见
// packages/@core/ui-kit/shadcn-ui/src/ui/tree/tree.vue flatten 函数），
// 所以这里直接平铺原对象即可，模板插槽可通过 value.value.* 访问原字段。
function transformToTreeData(items: any[]): TreeDataItem[] {
  return items.map((item) => ({
    key: item.id,
    title: item.title,
    // 兼容框架默认 labelField
    label: item.title,
    // action 类型不显示图标
    icon: item.type === PermissionTypeOptionsValueAction ? '' : item.icon,
    type:item.type,
    children: item.children ? transformToTreeData(item.children) : undefined,
  }))
}

// 树节点选中变化
function onTreeCheck(keys: (number | string)[], info: any) {
  checkedKeys.value = keys as string[]
  halfCheckedKeys.value = info?.halfCheckedKeys ?? []
  // 同步到 form 字段（如果 schema 启用了字段绑定）
  formApi.setFieldValue('menuPermission', checkedKeys.value)

  // 当用户选择了菜单时，清除错误状态
  if (checkedKeys.value.length > 0) {
    menuValidateStatus.value = ''
    menuHelpText.value = ''
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
    fieldName: '',
    component: 'FormTitle',
    label: '',
    labelWidth: 0,
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      title: $t('permission.role.form_group.basicInfo'),
    },
  },
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
  {
    fieldName: '',
    component: 'FormTitle',
    label: '',
    labelWidth: 0,
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      title: $t('permission.role.form_group.authorizedMenu'),
    },
  },
  {
    component: 'Input',
    fieldName: 'menuPermission',
    formItemClass: 'col-span-2 md:col-span-2 items-start',
    labelWidth: 0,
    modelPropName: 'modelValue',
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
  onOpenChange: async (isOpen) => {
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
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'">
      <template #menuPermission="slotProps">
        <div class="w-full">
          <Spin :spinning="menuTreeLoading" class="w-full">
            <Tree :tree-data="menuTreeData" multiple bordered class="w-full" :default-expanded-level="2"
              v-bind="slotProps" :model-value="checkedKeys" value-field="key" label-field="title" icon-field="icon"
              :get-node-class="getNodeClass" @update:model-value="onTreeCheck">
              <template #node="{ value }">
                <IconifyIcon v-if="value?.type !== PermissionTypeOptionsValueAction && value?.icon" :icon="value.icon"
                  class="mr-1 size-4" />
                {{ $te(value.title) ? $t(value.title) : (value.title ?? '') }}
              </template>
            </Tree>
          </Spin>
        </div>
      </template>
    </Form>
  </Modal>
</template>

<style lang="scss" scoped>
:deep(.ant-tree-title) {
  .tree-actions {
    @apply ml-5 hidden;
  }
}

:deep(.ant-tree-title:hover) {
  .tree-actions {
    @apply ml-5 flex flex-auto justify-end;
  }
}

// action 类型（按钮权限）样式：去除图标位置，用细体区分
:deep(.role-menu-action) {
  font-style: italic;
  color: hsl(var(--muted-foreground));
}
</style>