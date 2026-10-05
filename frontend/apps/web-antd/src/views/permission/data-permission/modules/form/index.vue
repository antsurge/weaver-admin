<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { SystemDataPermissionApi } from '#/api/permission/data-permission';
import type { PermissionRoleApi } from '#/api/permission/role';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import { getDepartmentTreeApi } from '#/api/organization/department';
import {
  createDataPermissionApi,
  isDataPermissionCodeExistsApi,
  updateDataPermissionApi,
} from '#/api/permission/data-permission';
import { getRoleListApi } from '#/api/permission/role';
import { $t } from '#/locales';

import { SCOPE_TYPE_OPTIONS } from '../../data';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<SystemDataPermissionApi.DataPermission>();

const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('permission.data_permission.fields.name'),
    component: 'Input',
    rules: z
      .string()
      .min(
        2,
        $t('ui.formRules.minLength', [
          $t('permission.data_permission.fields.name'),
          2,
        ]),
      )
      .max(
        50,
        $t('ui.formRules.maxLength', [
          $t('permission.data_permission.fields.name'),
          50,
        ]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('permission.data_permission.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'code',
    label: $t('permission.data_permission.fields.code'),
    component: 'Input',
    rules: z
      .string()
      .min(
        2,
        $t('ui.formRules.minLength', [
          $t('permission.data_permission.fields.code'),
          2,
        ]),
      )
      .max(
        50,
        $t('ui.formRules.maxLength', [
          $t('permission.data_permission.fields.code'),
          50,
        ]),
      )
      .refine(
        async (value: string) => {
          if (!value) {
            return true;
          }
          const res = await isDataPermissionCodeExistsApi(
            value,
            formData.value?.id,
          );
          return !res?.exists;
        },
        (value) => ({
          message: $t('ui.formRules.alreadyExists', [
            $t('permission.data_permission.fields.code'),
            value,
          ]),
        }),
      ),
    // 失焦时触发字段校验（请求接口判断编码是否存在）
    formFieldProps: { validateOnBlur: true },
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('permission.data_permission.fields.code'),
      ]),
    },
  },
  {
    fieldName: 'scopeType',
    label: $t('permission.data_permission.fields.scopeType'),
    component: 'Select',
    defaultValue: '1',
    rules: 'required',
    componentProps: {
      class: 'w-full',
      placeholder: $t('ui.formRules.selectRequired', [
        $t('permission.data_permission.fields.scopeType'),
      ]),
      options: SCOPE_TYPE_OPTIONS,
    },
  },
  {
    fieldName: 'deptIds',
    label: $t('permission.data_permission.fields.deptIds'),
    component: 'ApiTreeSelect',
    componentProps: {
      api: getDepartmentTreeApi,
      afterFetch: (res: AnyObject) => {
        const list = (res as any)?.items || (res as any);
        const convert = (items: any[]): any[] =>
          (items || []).map((item) => ({
            value: item.id,
            label: item.name,
            children: convert(item.children || []),
          }));
        return convert(list || []);
      },
      allowClear: true,
      class: 'w-full',
      multiple: true,
      placeholder: $t('permission.data_permission.form_placeholder.deptIds'),
      showSearch: true,
      treeDefaultExpandAll: true,
      getPopupContainer,
    },
    dependencies: {
      triggerFields: ['scopeType'],
      show: (values) => values.scopeType === '2',
    },
  },
  {
    fieldName: 'roleIds',
    label: $t('permission.data_permission.fields.roleIds'),
    component: 'ApiSelect',
    componentProps: {
      api: async () => {
        const res = await getRoleListApi({ currentPage: 1, pageSize: 100 });
        const list = res?.items || [];
        return list
          .filter((item) => item.status === 'enabled')
          .map((item: PermissionRoleApi.Role) => ({
            value: item.id,
            label: item.name,
          }));
      },
      allowClear: true,
      class: 'w-full',
      mode: 'multiple',
      optionFilterProp: 'label',
      placeholder: $t('permission.data_permission.form_placeholder.roleIds'),
      showSearch: true,
      getPopupContainer,
    },
    dependencies: {
      triggerFields: ['scopeType'],
      show: (values) => values.scopeType === '2',
    },
  },
  {
    fieldName: 'remark',
    label: $t('permission.data_permission.fields.remark'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('permission.data_permission.form_placeholder.remark'),
      autoSize: { minRows: 2, maxRows: 4 },
    },
  },
  {
    fieldName: 'status',
    label: $t('permission.data_permission.fields.status'),
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
];

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    colon: true,
    labelWidth: 80,
  },
  schema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-1',
});

/** JSON 字符串 -> 数组（编辑回填） */
function parseIdList(value?: string): string[] {
  if (!value) return [];
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) ? parsed.map(String) : [];
  } catch {
    return [];
  }
}

/** 数组 -> JSON 字符串（提交） */
function serializeIdList(value?: string | string[]): string | undefined {
  const list = Array.isArray(value) ? value : parseIdList(value as string);
  return list.length > 0 ? JSON.stringify(list) : undefined;
}

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange(isOpen) {
    if (!isOpen) return;
    const data = modalApi.getData<SystemDataPermissionApi.DataPermission>();
    if (data) {
      formData.value = data;
      formApi.setValues({
        ...data,
        deptIds: parseIdList(data.deptIds),
        roleIds: parseIdList(data.roleIds),
      });
    } else {
      formData.value = undefined;
      formApi.resetForm();
    }
  },
});

async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) return;

  modalApi.lock();
  try {
    const values =
      (await formApi.getValues()) as SystemDataPermissionApi.DataPermissionInput;

    const data: SystemDataPermissionApi.DataPermissionInput = {
      ...values,
      deptIds: serializeIdList(values.deptIds),
      roleIds: serializeIdList(values.roleIds),
    };

    await (formData.value?.id
      ? updateDataPermissionApi(formData.value.id, data)
      : createDataPermissionApi(data));

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('permission.data_permission.name')])
    : $t('ui.actionTitle.create', [$t('permission.data_permission.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[600px]" :title="getModalTitle">
    <Form :layout="isHorizontal ? 'horizontal' : 'vertical'" class="mx-4" />
  </Modal>
</template>
