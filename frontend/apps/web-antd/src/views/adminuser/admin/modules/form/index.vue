<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';

import { computed, onMounted, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { Select } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import {
  getAdminApi,
  createAdminApi,
  updateAdminApi,
} from '#/api/adminuser/admin';
import { getRoleListApi } from '#/api/permission/role';
import type { AdminuserAdminApi } from '#/api/adminuser/admin';
import type { PermissionRoleApi } from '#/api/permission/role';

import { $t } from '#/locales';

import { realNameRule, usernameRule } from './rules';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<AdminuserAdminApi.Admin>();

// 角色相关状态
const roleOptions = ref<{ label: string; value: string }[]>([]);
const selectedRoleIds = ref<string[]>([]);
const roleLoading = ref(false);

// 加载角色列表
async function loadRoleList() {
  roleLoading.value = true;
  try {
    const res = await getRoleListApi({ page: 1, pageSize: 1000 });
    // 转换为 Select 组件需要的格式
    roleOptions.value = (res.items || []).map((item: PermissionRoleApi.Role) => ({
      label: item.name,
      value: item.id,
    }));
  } catch (error) {
    console.error('加载角色列表失败:', error);
  } finally {
    roleLoading.value = false;
  }
}

 const schema: VbenFormSchema[] = [
  {
    fieldName: 'realName',
    label: '真实姓名',
    component: 'Input',
    rules: realNameRule,
    componentProps: {
      placeholder: '请输入真实姓名',
    },
  },
  {
    fieldName: 'username',
    label: '用户名',
    component: 'Input',
    rules: usernameRule,
    componentProps: {
      placeholder: '请输入用户名',
    },
  },
  {
    fieldName: 'email',
    label: '邮箱',
    component: 'Input',
    componentProps: {
      placeholder: '请输入邮箱（可选）',
    },
  },
  {
    fieldName: 'phone',
    label: '手机号',
    component: 'Input',
    componentProps: {
      placeholder: '请输入手机号（可选）',
    },
  },
  {
    fieldName: 'avatar',
    label: '头像',
    component: 'Input',
    componentProps: {
      placeholder: '请输入头像地址（可选）',
    },
  },
  {
    fieldName: 'status',
    label: '状态',
    component: 'Switch',
    defaultValue: 'enabled',
    componentProps: {
      class: 'w-auto',
      checkedChildren: '启用',
      checkedValue: 'enabled',
      unCheckedChildren: '禁用',
      unCheckedValue: 'disabled',
    },
  },
  {
    fieldName: 'password',
    label: '密码',
    component: 'VbenInputPassword',
    componentProps: {
      placeholder: '编辑时留空则不修改；创建时留空使用默认密码 123456',
      allowClear: true,
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

    // 加载角色列表（只在首次打开时加载）
    if (roleOptions.value.length === 0) {
      await loadRoleList();
    }

    const data = modalApi.getData<AdminuserAdminApi.Admin>();
    // 编辑
    if (data?.id) {
      modalApi.lock();
      try {
        const res = await getAdminApi(data.id);
        formData.value = res;
        formApi.setValues(res);
        // 回显已绑定的角色
        selectedRoleIds.value = res.roleIds || [];
      } finally {
        modalApi.unlock();
      }
    } else {
      // 👉 新增
      formData.value = undefined;
      formApi.resetForm();
      // 清空角色选择
      selectedRoleIds.value = [];
    }
  },
});

async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) return;
  modalApi.lock();
  const data = await formApi.getValues<AdminuserAdminApi.Admin>();
  try {
    // 合并选中的角色ID
    data.roleIds = selectedRoleIds.value;

    if (formData.value?.id) {
      await updateAdminApi(formData.value.id, data);
    } else {
      await createAdminApi(data);
    }
    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', ['用户'])
    : $t('ui.actionTitle.create', ['用户']),
);

// 组件挂载时加载角色列表
onMounted(() => {
  loadRoleList();
});

// 过滤函数
function filterOption(input: string, option: { label?: string; value: string }) {
  return option.label?.toLowerCase().includes(input.toLowerCase()) ?? false;
}
</script>

<template>
  <Modal class="w-full max-w-[720px]" :title="getModalTitle">
    <div class="mx-4">
      <!-- 基本信息表单 -->
      <Form class="mb-6" :layout="isHorizontal ? 'horizontal' : 'vertical'" />

      <!-- 角色选择（多选） -->
      <div class="border-t pt-4">
        <div class="text-base font-medium mb-3 text-gray-700 flex items-center gap-1">
          分配角色
        </div>
        <div v-if="roleLoading" class="flex justify-center py-4">
          <a-spin tip="加载角色列表..." />
        </div>
        <Select
          v-else
          v-model:value="selectedRoleIds"
          mode="multiple"
          placeholder="请选择角色"
          style="width: 100%"
          :options="roleOptions"
          :filter-option="filterOption"
          allow-clear
        />
      </div>
    </div>
  </Modal>
</template>
