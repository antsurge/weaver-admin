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

import { $t} from '#/locales';
import { $te } from '@vben/locales';

import {
  getPermissionTypeOptions,
  getBadgeTypeOptions,
  getBadgeStyleOptions,
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
      title: $t('permission.menu.form_group.basicInfo'),
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
      placeholder: $t('ui.formRules.selectRequired', [$t('permission.menu.fields.parentID')]),
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
    componentProps() {
      // 不需要处理多语言时就无需这么做
      return {
        placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.title")]),
        ...(titleSuffix.value && { addonAfter: titleSuffix.value }),
        onChange({ target: { value } }: { target: { value: string } }) {
          titleSuffix.value = value && $te(value) ? $t(value) : undefined;
        },
      };
    },
    description: $t('permission.menu.description.title'),
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
    component: 'EnhancedIconPicker',
    formItemClass: 'col-span-1 md:col-span-1',
    componentProps: () => ({
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.icon")]),
      // 启用的图标源（可自定义）
      sources: ['ant-design', 'bootstrap-icons', 'lucide', 'local-svg'],
    }),
    label: $t('permission.menu.fields.icon'),
    dependencies: {
      triggerFields: ['type'],
      show: (values) => values.type !== PermissionTypeOptionsValueAction,
    },
  },

  // ==================== 路由配置 ====================
   {
    fieldName: '',
    component: 'FormTitle',
    label: '',
    labelWidth: 0,
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      title: $t('permission.menu.form_group.routeConfiguration'),
    },
  },
  {
    fieldName: 'path',
    component: 'Input',
    label: $t('permission.menu.fields.path'),
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.path")])
    },
    description:$t("permission.menu.description.path"),
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return [PermissionTypeOptionsValueCatalog, PermissionTypeOptionsValueMenu]
          .includes(values.type)
      },
    },
    rules: pathRule,
  },
  {
     fieldName: 'component',
    component: 'Input',
    label: $t('permission.menu.fields.component'),
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.component")])
    },
    description: $t("permission.menu.description.component"),
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return [PermissionTypeOptionsValueMenu]
          .includes(values.type)
      },
    },
    rules: componentRule,
  },
  // ==================== 权限与状态 ====================
  {
    fieldName: '',
    component: 'FormTitle',
    label: '',
    labelWidth: 0,
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      title: $t('permission.menu.form_group.permissionAndStatus'),
    },
  },
  {
     fieldName: 'authCode',
    component: 'Input',
    label: $t('permission.menu.fields.authCode'),
    formItemClass: 'col-span-1 md:col-span-1',
    componentProps: {
      placeholder: $t("permission.menu.form_placeholder.authCode")
    },
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return [PermissionTypeOptionsValueCatalog,PermissionTypeOptionsValueMenu,PermissionTypeOptionsValueAction]
          .includes(values.type)
      },
    },
    rules: componentRule,
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
    formItemClass: 'col-span-1 md:col-span-1',
    defaultValue: 'enabled',
    label: $t('permission.menu.fields.status'),
  },
  
  // ==================== 徽标配置 ====================
  {
    fieldName: '',
    component: 'FormTitle',
    label: '',
    labelWidth: 0,
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      title: $t('permission.menu.form_group.badgeConfiguration'),
    },
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return values.type !== PermissionTypeOptionsValueAction;
      },
    },
  },
  {
    fieldName: 'badgeType',
    component: 'Select',
    componentProps: {
      allowClear: true,
      class: 'w-full',
      options: getBadgeTypeOptions(),
    },
    dependencies: {
      show: (values) => {
        return values.type !== PermissionTypeOptionsValueAction;
      },
      triggerFields: ['type'],
    },
    label: $t('permission.menu.fields.badgeType'),
    formItemClass: 'col-span-1 md:col-span-1',
  },
  {
    fieldName: 'badgeContent',
    component: 'Input',
    label: $t('permission.menu.fields.badgeContent'),
    formItemClass: 'col-span-1 md:col-span-1',
    componentProps: {
    },
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return values.type !== PermissionTypeOptionsValueAction;
      },
    },
  },
  {
    fieldName: 'badgeStyle',
    component: 'Select',
    componentProps: {
      allowClear: true,
      class: 'w-full',
      options: getBadgeStyleOptions(),
    },
    dependencies: {
      show: (values) => {
        return values.type !== PermissionTypeOptionsValueAction;
      },
      triggerFields: ['type'],
    },
    label: $t('permission.menu.fields.badgeStyle'),
    formItemClass: 'col-span-1 md:col-span-1',
  },

  // ==================== 其他设置 ====================
  {
    fieldName: '',
    component: 'FormTitle',
    label: '',
    labelWidth: 0,
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      title: $t('permission.menu.form_group.otherConfiguration'),
    },
  },
  {
    fieldName: 'remark',
    component: 'Textarea',
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.remark")])
    },
    label: $t('permission.menu.fields.remark'),
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
