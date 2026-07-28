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
  titleRule,
  componentRule,
  pathRule,
  authCodeRule,
  linkUrlRule,
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
  getBadgeVariantsOptions,
  PermissionTypeOptionsValueMenu,
  // PermissionTypeOptionsValueTab,
  PermissionTypeOptionsValueLink,
  PermissionTypeOptionsValueIframe,
  PermissionTypeOptionsValueCatalog,
  PermissionTypeOptionsValueAction,
  PermissionBadgeTypeOptionsValueText,
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
  // ==================== 基础信息 ====================
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
            label: $t(item.title),
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
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.name")]),
      showCount: true,
      maxlength: 30,
    }
  },
  {
    fieldName: 'title',
    label: $t('permission.menu.fields.title'),
    component: 'Input',
    formItemClass: 'col-span-1 md:col-span-1',
    rules: titleRule,
    componentProps() {
      // 不需要处理多语言时就无需这么做
      return {
        placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.title")]),
        ...(titleSuffix.value && { addonAfter: titleSuffix.value }),
        onChange({ target: { value } }: { target: { value: string } }) {
          titleSuffix.value = value && $te(value) ? $t(value) : undefined;
        },
        showCount: true,
        maxlength: 60,
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
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return [PermissionTypeOptionsValueCatalog, 
        PermissionTypeOptionsValueMenu,PermissionTypeOptionsValueIframe,PermissionTypeOptionsValueLink]
          .includes(values.type)
      },
    },
  },
  {
    fieldName: 'path',
    component: 'Input',
    label: $t('permission.menu.fields.path'),
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.path")]),
      showCount: true,
      maxlength: 100,
    },
    description:$t("permission.menu.description.path"),
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return [PermissionTypeOptionsValueCatalog, PermissionTypeOptionsValueMenu,PermissionTypeOptionsValueIframe]
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
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.component")]),
      showCount: true,
      maxlength: 100,
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
  {
    fieldName: 'linkUrl',
    component: 'Input',
    label: $t('permission.menu.fields.linkUrl'),
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.linkUrl")]),
      showCount: true,
      maxlength: 100,
    },
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return [PermissionTypeOptionsValueIframe,PermissionTypeOptionsValueLink]
          .includes(values.type)
      },
    },
    rules: linkUrlRule,
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
      placeholder:  $t("ui.formRules.required", [$t("permission.menu.fields.authCode")]),
      showCount: true,
      maxlength: 100,
    },
    dependencies: {
      triggerFields: ['type'],
      show: (values) => {
        return [PermissionTypeOptionsValueCatalog,PermissionTypeOptionsValueMenu,PermissionTypeOptionsValueAction]
          .includes(values.type)
      },
    },
    rules: authCodeRule,
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
      placeholder: $t("ui.formRules.selectRequired", [$t("permission.menu.fields.badgeType")])
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
    fieldName: 'badge',
    component: 'Input',
    label: $t('permission.menu.fields.badge'),
    formItemClass: 'col-span-1 md:col-span-1',
    componentProps: {
      placeholder: $t("ui.formRules.required", [$t("permission.menu.fields.badge")]),
      showCount: true,
      maxlength: 10,
    },
    dependencies: {
      triggerFields: ['type', 'badgeType'],
      show: (values) => {
        return values.type !== PermissionTypeOptionsValueAction;
      },
      // 只有 badgeType 选择 dot 时才能输入
      disabled: (values) => values.badgeType !== PermissionBadgeTypeOptionsValueText,
      // badgeType 不是 text 时，清空 badge 的值
      trigger: (values, formApi) => {
        if (values.badgeType !== PermissionBadgeTypeOptionsValueText && values.badge) {
          formApi.setFieldValue('badge', undefined);
        }
      },
    },
  },
  {
    fieldName: 'badgeVariants',
    component: 'Select',
    componentProps: {
      allowClear: true,
      class: 'w-full',
      options: getBadgeVariantsOptions(),
      placeholder: $t("ui.formRules.selectRequired", [$t("permission.menu.fields.badgeVariants")])
    },
    dependencies: {
      show: (values) => {
        return values.type !== PermissionTypeOptionsValueAction;
      },
      triggerFields: ['type'],
    },
    label: $t('permission.menu.fields.badgeVariants'),
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

   {
    fieldName: '',
    component: 'FormTitle',
    label: '',
    labelWidth: 0,
    formItemClass: 'col-span-2 md:col-span-2',
    componentProps: {
      title: $t('permission.menu.form_group.apiPermissions'),
    },
    dependencies: {
      triggerFields: ['type'],
      show: (values) => values.type === PermissionTypeOptionsValueAction,
    },
  },
  // 接口权限（仅按钮类型）
  {
    fieldName: 'apiPermissions',
    component: 'ApiPermissionPicker',
    labelWidth:0,
    formItemClass: 'col-span-2 md:col-span-2',
    dependencies: {
      triggerFields: ['type'],
      show: (values) => values.type === PermissionTypeOptionsValueAction,
    },
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

      // title 是顶级字段（后端 proto/ent 均为顶级），从 data.title 读取 i18n key
      const titleVal = data.title as string;
      titleSuffix.value = titleVal && $te(titleVal) ? $t(titleVal) : '';
    } else {
      formApi.resetForm();
      titleSuffix.value = '';
    }
  },
});

/**
 * 根据菜单类型清理表单中隐藏的字段
 */
function cleanHiddenFields(data: Record<string, any>) {
  switch (data.type) {
    case PermissionTypeOptionsValueCatalog:
      // 目录：隐藏 component、linkUrl
      delete data.component;
      delete data.linkUrl;
      break;
    case PermissionTypeOptionsValueMenu:
      // 菜单：隐藏 linkUrl
      delete data.linkUrl;
      break;
    case PermissionTypeOptionsValueAction:
      // 按钮：隐藏 icon、path、component、linkUrl 及徽标相关
      delete data.icon;
      delete data.path;
      delete data.component;
      delete data.linkUrl;
      delete data.badgeType;
      delete data.badge;
      delete data.badgeVariants;
      break;
    case PermissionTypeOptionsValueIframe:
      // iframe：隐藏 component、authCode
      delete data.component;
      delete data.authCode;
      break;
    case PermissionTypeOptionsValueLink:
      // 外链：隐藏 path、component、authCode
      delete data.path;
      delete data.component;
      delete data.authCode;
      break;
  }
}

/**
 * 提交
 */
async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) return;

  modalApi.lock();

  const data = await formApi.getValues();

  // 清理隐藏字段：根据 type 删除 schema 中通过 dependencies.show 隐藏的字段
  cleanHiddenFields(data);

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
