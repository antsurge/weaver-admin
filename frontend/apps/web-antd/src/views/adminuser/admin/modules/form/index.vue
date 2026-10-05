<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { AdminuserAdminApi } from '#/api/adminuser/admin';
import type { PermissionRoleApi } from '#/api/permission/role';

import { computed, onMounted, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';
import { message, Select } from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import {
  createAdminApi,
  getAdminApi,
  isUsernameExistsApi,
  updateAdminApi,
} from '#/api/adminuser/admin';
import { getDepartmentTreeApi } from '#/api/organization/department';
import { getDataPermissionListApi } from '#/api/permission/data-permission';
import { getRoleListApi } from '#/api/permission/role';
import { getFileAccessUrl, uploadFileApi } from '#/api/system/file';
import { $t } from '#/locales';

import { emailRule, passwordRule, phoneRule, realNameRule } from './rules';

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
    const res = await getRoleListApi({ currentPage: 1, pageSize: 1000 });
    // 转换为 Select 组件需要的格式
    roleOptions.value = (res.items || []).map(
      (item: PermissionRoleApi.Role) => ({
        label: item.name,
        value: item.id,
      }),
    );
  } catch (error) {
    console.error('加载角色列表失败:', error);
  } finally {
    roleLoading.value = false;
  }
}

// ─── 头像上传 ────────────────────────────────────────────
/** 头像上传目录（相对 upload_prefix） */
const AVATAR_DIR = 'avatar';

/** 头像文件列表项 */
interface AvatarFileItem {
  uid: string;
  name: string;
  status: string;
  url: string;
  /** 已回显文件的来源对象键，用于未重新上传时原样提交 */
  objectKey?: string;
}

/** 将后端返回的头像对象键转为 Upload 组件需要的文件列表 */
function avatarKeyToFileList(
  objectKey?: string,
  accessUrl?: string,
): AvatarFileItem[] {
  if (!objectKey) return [];
  return [
    {
      uid: `avatar-${Date.now()}`,
      name: objectKey.split('/').pop() || 'avatar',
      status: 'done',
      // 优先使用后端返回的完整访问地址，回退到前端拼接
      url: accessUrl || getFileAccessUrl(objectKey),
      objectKey,
    },
  ];
}

/** 从 Upload 组件的文件列表中提取头像对象键（存入后端） */
function fileListToAvatarKey(fileList?: any): string | undefined {
  let list: any[];
  if (Array.isArray(fileList)) {
    list = fileList;
  } else if (fileList) {
    list = [fileList];
  } else {
    list = [];
  }
  const done = list.find((f: any) => f?.status === 'done');
  // 新上传：response 为上传接口返回结果，取 objectKey；
  // 已回显未重传：取回显时保留的 objectKey
  return done?.response?.objectKey || done?.objectKey || undefined;
}

/** 将裁剪后的 Blob 包装成带扩展名的 File，避免后端扩展名校验失败 */
function toNamedFile(file: Blob | File): Blob | File {
  if (file instanceof File && file.name) return file;
  const ext = (file.type || '').includes('png') ? '.png' : '.jpg';
  return new File([file], `avatar_${Date.now()}${ext}`, {
    type: file.type || 'image/jpeg',
  });
}

/** 自定义上传：调用对象存储上传接口 */
async function handleAvatarUpload(options: any) {
  const { file, onError, onSuccess } = options;
  try {
    const res = await uploadFileApi({
      file: toNamedFile(file),
      dir: AVATAR_DIR,
    });
    // 上传返回的直链在私有桶下不可直接访问，统一走带鉴权的访问地址
    onSuccess?.({ ...res, url: getFileAccessUrl(res.objectKey) });
  } catch (error) {
    message.error('头像上传失败');
    onError?.(error as Error);
  }
}
// ────────────────────────────────────────────────────────

const schema: VbenFormSchema[] = [
  {
    fieldName: 'departmentId',
    label: '归属部门',
    component: 'ApiTreeSelect',
    rules: z.string().optional(),
    componentProps: {
      api: getDepartmentTreeApi,
      allowClear: true,
      class: 'w-full',
      showSearch: true,
      treeDefaultExpandAll: true,
      labelField: 'label',
      valueField: 'id',
      childrenField: 'children',
      getPopupContainer,
      placeholder: '请选择归属部门，留空表示不归属',
      afterFetch: (res: any[]) => {
        const list = (res as any)?.items || res;
        const convert = (items: any[]): any[] =>
          items.map((item) => ({
            id: item.id,
            label: item.name,
            children: convert(item.children || []),
          }));
        return convert(list || []);
      },
    },
  },
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
    rules: z
      .string()
      .min(
        2,
        $t('ui.formRules.minLength', [
          $t('adminuser.admin.fields.username'),
          2,
        ]),
      )
      .max(
        30,
        $t('ui.formRules.maxLength', [
          $t('adminuser.admin.fields.username'),
          30,
        ]),
      )
      .refine(
        async (value: string) => {
          if (!value) {
            return false;
          }
          const res = await isUsernameExistsApi(value, formData.value?.id);
          return !res?.exists;
        },
        (value) => ({
          message: $t('ui.formRules.alreadyExists', [
            $t('adminuser.admin.fields.username'),
            value,
          ]),
        }),
      ),
    componentProps: {
      placeholder: '请输入用户名',
    },
    // 失焦时校验用户名是否已存在
    formFieldProps: { validateOnBlur: true },
  },
  {
    fieldName: 'email',
    label: '邮箱',
    component: 'Input',
    rules: emailRule,
    componentProps: {
      placeholder: '请输入邮箱（可选）',
    },
  },
  {
    fieldName: 'phone',
    label: '手机号',
    component: 'Input',
    rules: phoneRule,
    componentProps: {
      placeholder: '请输入手机号（可选）',
    },
  },
  {
    fieldName: 'avatar',
    label: '头像',
    component: 'Upload',
    componentProps: {
      // 单张图片上传，带裁剪与预览
      listType: 'picture-card',
      maxCount: 1,
      accept: 'image/*',
      crop: true,
      aspectRatio: '1/1',
      // 自定义上传：走对象存储上传接口
      customRequest: handleAvatarUpload,
    },
    help: '支持 jpg/png/gif/webp 等图片格式，建议 1:1 正方形',
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
    fieldName: 'dataPermissionIds',
    label: $t('adminuser.admin.fields.dataPermissionIds'),
    component: 'ApiSelect',
    componentProps: {
      api: async () => {
        const res = await getDataPermissionListApi({
          status: 'enabled',
          currentPage: 1,
          pageSize: 200,
        });
        const list = res?.items || [];
        return list.map((item) => ({
          value: item.id,
          label: `${item.name}（${item.code}）`,
        }));
      },
      allowClear: true,
      class: 'w-full',
      getPopupContainer,
      mode: 'multiple',
      multiple: true,
      optionFilterProp: 'label',
      placeholder: $t('adminuser.admin.form_placeholder.dataPermissionIds'),
      showSearch: true,
    },
    help: $t('adminuser.admin.form_help.dataPermissionIds'),
  },
  {
    fieldName: 'password',
    label: '密码',
    component: 'VbenInputPassword',
    rules: passwordRule,
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
  onOpenChange: async (isOpen) => {
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
        // 头像：对象键 → 文件列表，供 Upload 组件回显（优先用后端返回的完整地址）
        formApi.setValues({
          ...res,
          avatar: avatarKeyToFileList(res.avatar, res.avatarUrl),
        });
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
      // 左侧选中部门时，默认选中该部门（可在树中改选）
      if (data?.departmentId) {
        formApi.setValues({
          departmentId: data.departmentId,
        });
      }
    }
  },
});

async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) return;
  modalApi.lock();
  const data = await formApi.getValues<AdminuserAdminApi.Admin>();
  try {
    // 头像：文件列表 → 对象键，供后端存储
    data.avatar = fileListToAvatarKey((data as any).avatar);
    // 合并选中的角色ID
    data.roleIds = selectedRoleIds.value;

    await (formData.value?.id
      ? updateAdminApi(formData.value.id, data)
      : createAdminApi(data));
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
function filterOption(input: string, option: any) {
  return option?.label?.toLowerCase().includes(input.toLowerCase()) ?? false;
}
</script>

<template>
  <Modal class="w-full max-w-[720px]" :title="getModalTitle">
    <div class="mx-4">
      <!-- 基本信息表单 -->
      <Form class="mb-6" :layout="isHorizontal ? 'horizontal' : 'vertical'" />

      <!-- 角色选择（多选） -->
      <div class="border-t pt-4">
        <div
          class="mb-3 flex items-center gap-1 text-base font-medium text-gray-700"
        >
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
