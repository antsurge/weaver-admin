<script lang="ts" setup>
import type { ChangeEvent } from 'ant-design-vue/es/_util/EventInterface';
import type { Recordable } from '@vben/types';
import type { VbenFormSchema } from '#/adapter/form';

import { computed, h, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { getPopupContainer } from '@vben/utils';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';
import type { AllResult } from '#/types/pagination'

import { useVbenForm } from '#/adapter/form';

import {
  nameRule,
  codeRule,
  componentRule,
  urlRule,
  pathRule,
} from "#/views/permission/menu/modules/form/rules"

import {
  createMenuApi,
  getMenuTreeApi,
  updateMenuApi,
} from '#/api/permission/menu';
import type {
  PermissionMenuApi,
} from '#/api/permission/menu';

import { $t } from '#/locales';
import {
  getPermissionTypeOptions,
  getPermissionMenuTypeOptions,
  PermissionTypeOptionsValueMenu,
  // PermissionTypeOptionsValueTab,
  PermissionTypeOptionsValueLink,
  PermissionTypeOptionsValueIframe,
  PermissionTypeOptionsValueCatalog,
  PermissionTypeOptionsValueAction,
} from '#/views/permission/menu/data';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<PermissionMenuApi.PermissionMenu>();
const titleSuffix = ref<string>();

/**
 * 表单 schema
 */
const schema: VbenFormSchema[] = [
  // ==================== 基础设置 ====================
  {
    fieldName: '',
    component: 'FormTitle',
    label: '',
    labelWidth: 0,
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      title: '基础设置',
    },
  },
  // 类型 - 占据整行
  {
    fieldName: 'type',
    label: $t('permission.menu.fields.type'),
    defaultValue: PermissionTypeOptionsValueCatalog,
    component: 'RadioGroup',
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      optionType: "button",
      options: getPermissionTypeOptions(),
    },
  },
  // 上级菜单 - 占据整行
  {
    fieldName: 'parentID',
    label: $t('permission.menu.fields.parentID'),
    component: 'ApiTreeSelect',
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      api: getMenuTreeApi,
      allowClear: true,
      afterFetch: (res:AllResult<PermissionMenuApi.PermissionMenu>) => {
        const list = res?.items || res;
        const convert = (list: any[]): any[] =>
          list.map((item) => ({
            id: item.id,
            label: item.name,
            icon: item.icon,
            children: convert(item.children || []),
          }));

        return convert(list);
      },
      class: "w-full",
      showSearch: true,
      treeDefaultExpandAll: true,
      labelField: 'label',
      valueField: 'id',
      childrenField: 'children',
      getPopupContainer,
      placeholder: $t('permission.menu.form_placeholder.parentID')
    },
    renderComponentContent() {
      return {
        title({ label, meta }: { label: string; meta: Recordable<any> }) {
          const nodes = [];
          if (meta?.icon) {
            nodes.push(h(IconifyIcon, { class: 'size-4', icon: meta.icon }));
          }
          nodes.push(h('span', {}, $t(label || '')));
          return h('div', { class: 'flex items-center gap-1' }, nodes);
        },
      };
    },
  },
  // 菜单名称 + 菜单编码（一行两个字段）
  {
    fieldName: 'name',
    label: $t('permission.menu.fields.name'),
    component: 'Input',
    formItemClass: 'col-span-1 md:col-span-1',
    rules: nameRule,
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.name")])
    }
  },
  {
    fieldName: 'title',
    label: $t('permission.menu.fields.title'),
    component: 'Input',
    formItemClass: 'col-span-1 md:col-span-1',
    rules: codeRule,
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.title")])
    }
  },
  // 权重 + 图标（一行两个字段，按钮类型时隐藏图标）
  {
    component: 'InputNumber',
    fieldName: 'weight',
    label: $t('permission.menu.fields.weight'),
    formItemClass: 'col-span-1 md:col-span-1',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.weight")]),
      min: 0,
      step: 1,
      precision: 0,
      stringMode: false,
      style: { width: '100%' },
      onInput: (e: Event) => {
        const input = e.target as HTMLInputElement;
        input.value = input.value.replace(/[^\d]/g, '');
      },
    },
  },
  {
    fieldName: 'icon',
    component: 'IconPicker',
    formItemClass: 'col-span-1 md:col-span-1',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.icon")]),
      disabled: true,
    },
    label: $t('permission.menu.fields.icon'),
    dependencies: {
      triggerFields: ['type'],
      show: (values) => values.type !== PermissionTypeOptionsValueAction,
    }
  },

  // ==================== 其他设置（保留原有逻辑）====================
  {
    fieldName: 'path',
    label: $t('permission.menu.fields.path'),
    component: 'Input',
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.path")])
    },
    dependencies: {
      triggerFields: ['type'],
      show: (values) => values.type !== PermissionTypeOptionsValueAction,
    },
    rules: pathRule,
  },
  {
    fieldName: 'menuType',
    component: 'RadioGroup',
    label: $t('permission.menu.fields.menuType'),
    formItemClass: 'col-span-2 md:col-span-2',
    defaultValue: 'tab',
    componentProps: {
      optionType: "button",
      options: getPermissionMenuTypeOptions(),
    },
    dependencies: {
      triggerFields: ['type'],
      show: (values) => values.type === PermissionTypeOptionsValueMenu,
    },
  },
  {
    fieldName: 'component',
    component: 'Input',
    label: $t('permission.menu.fields.component'),
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.placeholder.component")
    },
    dependencies: {
      triggerFields: ['type', 'menuType'],
      show: (values) => {
        return (values.type === PermissionTypeOptionsValueMenu && values.menuType === PermissionTypeMenuTypeOptionsValueTab)
      }
    },
    rules: componentRule,
  },
  {
    fieldName: 'url',
    component: 'Input',
    label: $t('permission.menu.fields.url'),
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.placeholder.url")
    },
    dependencies: {
      triggerFields: ['type', 'menuType'],
      show: (values) => {
        return (values.type === PermissionTypeOptionsValueMenu && [PermissionTypeMenuTypeOptionsValueLink, PermissionTypeMenuTypeOptionsValueIframe]
          .includes(values.menuType))
      },
    },
    rules: urlRule,
  },
  {
    fieldName: 'description',
    component: 'Textarea',
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.remark")])
    },
    label: $t('permission.menu.fields.remark'),
  },
  {
    component: 'Switch',
    componentProps: {
      class: 'w-auto',
      checkedChildren: $t('common.enabled'),
      checkedValue: "enabled",
      unCheckedChildren: $t('common.disabled'),
      unCheckedValue: "disabled",
    },
    fieldName: 'status',
    formItemClass: 'col-span-2 md:col-span-2',
    defaultValue: 'enabled',
    label: $t('permission.menu.fields.status'),
  },
];

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

/**
 * Form
 */
const [Form, formApi] = useVbenForm({
  commonConfig: {
    colon: true,
    formItemClass: 'col-span-2 md:col-span-2',
    labelWidth: 75,
  },
  schema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2 gap-x-0',
});

/**
 * Modal
 */
const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange(isOpen) {
    if (!isOpen) return;

    const data = modalApi.getData<PermissionMenuApi.PermissionMenu>();

    if (data) {
      formData.value = data;
      formApi.setValues(data);

      titleSuffix.value = data.meta?.title ? $t(data.meta.title) : '';
    } else {
      formApi.resetForm();
      titleSuffix.value = '';
    }
  },
});

/**
 * 提交
 */
async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) return;

  modalApi.lock();

  const data = await formApi.getValues();

  // 清理隐藏字段
  if (data.type === PermissionTypeOptionsValueAction) {
    delete data.path;
    delete data.icon;
    delete data.menuType;
    delete data.component;
    delete data.url;
  }

  // if (data.menuType !== PermissionTypeOptionsValueTab) {
  //   delete data.component;
  // }

  if (![PermissionTypeOptionsValueLink, PermissionTypeOptionsValueIframe]
    .includes(data.menuType)) {
    delete data.url;
  }

  try {
    if (formData.value?.id) {
      await updateMenuApi(formData.value.id, data);
    } else {
      await createMenuApi(data);
    }

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

/**
 * 标题
 */
const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('permission.menu.name')])
    : $t('ui.actionTitle.create', [$t('permission.menu.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[800px]" :title="getModalTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'" />
  </Modal>
</template>
