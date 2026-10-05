<script setup lang="ts">
import type { UploadFile } from 'ant-design-vue';

import { ref, watch } from 'vue';

import { useUserStore } from '@vben/stores';

import { Button, Input, message, Modal, Tabs, Upload } from 'ant-design-vue';

import { getUserInfoApi } from '#/api';
import { getFileAccessUrl, uploadFileApi } from '#/api/system/file';
import {
  updateCurrentUserApi,
  updateCurrentUserPasswordApi,
} from '#/api/system/profile';

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  'update:open': [value: boolean];
}>();

const userStore = useUserStore();

const AVATAR_DIR = 'avatar';

const activeTab = ref<'basic' | 'password'>('basic');

// ── 资料 ──────────────────────────────────────────────
const nickname = ref('');
const avatarFileList = ref<UploadFile[]>([]);
/** 当前头像对象键（回显原始值，未重传时原样提交） */
const currentAvatarKey = ref<string | undefined>(undefined);
/** 归属部门名称（只读展示） */
const departmentName = ref('');
/** 角色名称列表（只读展示） */
const roleNames = ref<string[]>([]);
const infoLoading = ref(false);
const savingInfo = ref(false);

/** 头像对象键 → Upload 文件列表（回显用） */
function avatarKeyToFileList(objectKey?: string): UploadFile[] {
  currentAvatarKey.value = objectKey;
  if (!objectKey) return [];
  return [
    {
      uid: `avatar-${Date.now()}`,
      name: objectKey.split('/').pop() || 'avatar',
      status: 'done',
      url: getFileAccessUrl(objectKey),
    },
  ];
}

/** Upload 文件列表 → 头像对象键（提交用） */
function fileListToAvatarKey(fileList: UploadFile[]): string | undefined {
  const done = fileList.find((f) => f?.status === 'done');
  // 新上传的取上传接口返回的 objectKey；未重传的回退到回显时的原始值
  return (done as any)?.response?.objectKey || currentAvatarKey.value;
}

/** 将裁剪后的 Blob 包装成带扩展名的 File，避免后端扩展名校验失败 */
function toNamedFile(file: Blob | File): Blob | File {
  if (file instanceof File && file.name) return file;
  const ext = file.type === 'image/png' ? 'png' : 'jpg';
  return new File([file], `avatar_${Date.now()}.${ext}`, {
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

async function loadUserInfo() {
  const data = await getUserInfoApi();
  userStore.setUserInfo(data);
  nickname.value = data.realName ?? '';
  avatarFileList.value = avatarKeyToFileList(data.avatar);
  departmentName.value = data.departmentName ?? '';
  roleNames.value = data.roleNames ?? [];
}

async function handleSaveInfo() {
  const avatar = fileListToAvatarKey(avatarFileList.value);
  if (!nickname.value.trim() && !avatar) {
    message.warning('昵称或头像至少填写一项');
    return;
  }
  savingInfo.value = true;
  try {
    await updateCurrentUserApi({
      realName: nickname.value.trim(),
      avatar,
    });
    // 刷新全局用户信息（头像 / 昵称在顶栏即时更新）
    await loadUserInfo();
    // 默认成功提示（接口自带），成功后关闭弹框
    emit('update:open', false);
  } finally {
    savingInfo.value = false;
  }
}

// ── 密码 ──────────────────────────────────────────────
const oldPassword = ref('');
const newPassword = ref('');
const confirmPassword = ref('');
const savingPassword = ref(false);

async function handleSavePassword() {
  if (!oldPassword.value) {
    message.warning('请输入旧密码');
    return;
  }
  if (!newPassword.value || newPassword.value.length < 6) {
    message.warning('新密码长度不能少于6位');
    return;
  }
  if (newPassword.value !== confirmPassword.value) {
    message.warning('两次输入的新密码不一致');
    return;
  }
  savingPassword.value = true;
  try {
    await updateCurrentUserPasswordApi({
      oldPassword: oldPassword.value,
      newPassword: newPassword.value,
    });
    message.success('密码修改成功');
    oldPassword.value = '';
    newPassword.value = '';
    confirmPassword.value = '';
  } finally {
    savingPassword.value = false;
  }
}

// 每次打开时刷新资料、重置密码表单
watch(
  () => props.open,
  async (open) => {
    if (!open) return;
    activeTab.value = 'basic';
    oldPassword.value = '';
    newPassword.value = '';
    confirmPassword.value = '';
    infoLoading.value = true;
    try {
      await loadUserInfo();
    } catch {
      // 拉取失败不阻塞弹框
    } finally {
      infoLoading.value = false;
    }
  },
);
</script>

<template>
  <Modal
    :open="open"
    title="个人中心"
    :width="520"
    :footer="false"
    :mask-closable="false"
    @update:open="(v: boolean) => emit('update:open', v)"
  >
    <Tabs v-model:active-key="activeTab">
      <Tabs.TabPane key="basic" tab="基本资料">
        <div class="flex flex-col gap-4 px-1 py-2">
          <div class="flex items-center gap-4">
            <span
              class="w-16 shrink-0 text-sm text-gray-600 dark:text-gray-300"
            >
              头像
            </span>
            <Upload
              :file-list="avatarFileList"
              :custom-request="handleAvatarUpload"
              list-type="picture-card"
              accept="image/*"
              :max-count="1"
              @update:file-list="(v: UploadFile[]) => (avatarFileList = v)"
            >
              <div
                class="flex flex-col items-center justify-center text-xs text-gray-400"
              >
                <span>点击上传</span>
                <span>支持裁剪 1:1</span>
              </div>
            </Upload>
          </div>
          <div class="flex items-center gap-4">
            <span
              class="w-16 shrink-0 text-sm text-gray-600 dark:text-gray-300"
            >
              昵称
            </span>
            <Input
              v-model:value="nickname"
              class="flex-1"
              placeholder="请输入昵称"
              :maxlength="30"
            />
          </div>
          <div class="flex items-center gap-4">
            <span
              class="w-16 shrink-0 text-sm text-gray-600 dark:text-gray-300"
            >
              所属部门
            </span>
            <span class="flex-1 text-sm text-gray-800 dark:text-gray-200">
              {{ departmentName || '-' }}
            </span>
          </div>
          <div class="flex items-center gap-4">
            <span
              class="w-16 shrink-0 text-sm text-gray-600 dark:text-gray-300"
            >
              角色
            </span>
            <span class="flex-1 text-sm text-gray-800 dark:text-gray-200">
              {{ roleNames.length > 0 ? roleNames.join('、') : '-' }}
            </span>
          </div>
          <div class="flex justify-end">
            <Button
              type="primary"
              :loading="infoLoading || savingInfo"
              :disabled="infoLoading"
              @click="handleSaveInfo"
            >
              保存资料
            </Button>
          </div>
        </div>
      </Tabs.TabPane>
      <Tabs.TabPane key="password" tab="修改密码">
        <div class="flex flex-col gap-4 px-1 py-2">
          <div class="flex items-center gap-4">
            <span
              class="w-16 shrink-0 text-sm text-gray-600 dark:text-gray-300"
            >
              旧密码
            </span>
            <Input.Password
              v-model:value="oldPassword"
              class="flex-1"
              placeholder="请输入旧密码"
            />
          </div>
          <div class="flex items-center gap-4">
            <span
              class="w-16 shrink-0 text-sm text-gray-600 dark:text-gray-300"
            >
              新密码
            </span>
            <Input.Password
              v-model:value="newPassword"
              class="flex-1"
              placeholder="请输入新密码（至少6位）"
            />
          </div>
          <div class="flex items-center gap-4">
            <span
              class="w-16 shrink-0 text-sm text-gray-600 dark:text-gray-300"
            >
              确认密码
            </span>
            <Input.Password
              v-model:value="confirmPassword"
              class="flex-1"
              placeholder="请再次输入新密码"
            />
          </div>
          <div class="flex justify-end">
            <Button
              type="primary"
              :loading="savingPassword"
              @click="handleSavePassword"
            >
              保存密码
            </Button>
          </div>
        </div>
      </Tabs.TabPane>
    </Tabs>
  </Modal>
</template>
