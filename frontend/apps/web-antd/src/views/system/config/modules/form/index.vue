<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { SystemConfigApi } from '#/api/system/config';

import { computed, h, onMounted, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';
import { Modal as AntModal, Divider, Input, message } from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import {
  createConfigApi,
  getConfigListApi,
  updateConfigApi,
} from '#/api/system/config';
import { $t } from '#/locales';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<SystemConfigApi.Config>();

// 已有分组选项（支持输入新分组 + 选择已有分组）
const groupOptions = ref<{ label: string; value: string }[]>([]);

async function loadGroupOptions() {
  try {
    const res = await getConfigListApi({ currentPage: 1, pageSize: 0 });
    const groups = new Set<string>();
    (res?.items ?? []).forEach((item) => {
      if (item.group) groups.add(item.group);
    });
    groupOptions.value = [...groups].map((g) => ({ label: g, value: g }));
  } catch {
    groupOptions.value = [];
  }
}

onMounted(loadGroupOptions);

// 新建分组：弹出输入框，创建后加入选项并自动选中
function handleCreateGroup() {
  let groupName = '';
  AntModal.confirm({
    title: $t('system.config.operation.createGroup'),
    icon: h(IconifyIcon, { icon: 'ant-design:plus-outlined' }),
    content: h(Input, {
      placeholder: $t('system.config.form_placeholder.group'),
      onChange: (e: Event) => {
        groupName = (e.target as HTMLInputElement).value.trim();
      },
    }),
    onOk() {
      if (!groupName) {
        message.warning($t('system.config.message.groupRequired'));
        return Promise.reject(
          new Error($t('system.config.message.groupRequired')),
        );
      }
      if (!groupOptions.value.some((g) => g.value === groupName)) {
        groupOptions.value.push({ label: groupName, value: groupName });
      }
      formApi.setValues({ group: groupName });
      return Promise.resolve();
    },
  });
}

const schema: VbenFormSchema[] = [
  {
    fieldName: 'name',
    label: $t('system.config.fields.name'),
    component: 'Input',
    rules: z
      .string()
      .min(
        2,
        $t('ui.formRules.minLength', [$t('system.config.fields.name'), 2]),
      )
      .max(
        50,
        $t('ui.formRules.maxLength', [$t('system.config.fields.name'), 50]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.config.fields.name'),
      ]),
    },
  },
  {
    fieldName: 'key',
    label: $t('system.config.fields.key'),
    component: 'Input',
    rules: z
      .string()
      .min(2, $t('ui.formRules.minLength', [$t('system.config.fields.key'), 2]))
      .max(
        100,
        $t('ui.formRules.maxLength', [$t('system.config.fields.key'), 100]),
      ),
    componentProps: {
      placeholder: $t('ui.formRules.required', [
        $t('system.config.fields.key'),
      ]),
    },
  },
  {
    fieldName: 'value',
    label: $t('system.config.fields.value'),
    component: 'Input',
    rules: z
      .string()
      .min(1, $t('ui.formRules.required', [$t('system.config.fields.value')])),
    componentProps: {
      placeholder: $t('system.config.form_placeholder.value'),
    },
  },
  {
    fieldName: 'group',
    label: $t('system.config.fields.group'),
    component: 'Select',
    componentProps: () => ({
      placeholder: $t('system.config.form_placeholder.group'),
      allowClear: true,
      showSearch: true,
      optionFilterProp: 'label',
      options: groupOptions.value,
      // antd Select 默认按内容自适应宽度，这里铺满整行
      class: 'w-full',
    }),
    renderComponentContent() {
      return {
        // antd 的 dropdownRender 参数为 { menuNode, props }，需解构出 menuNode VNode
        dropdownRender({ menuNode }: { menuNode: any }) {
          return h('div', {}, [
            menuNode,
            h(Divider, { style: { margin: '4px 0' } }),
            h(
              'div',
              {
                class: 'cursor-pointer px-3 py-1 text-primary',
                onClick: handleCreateGroup,
              },
              [
                h(IconifyIcon, { icon: 'ant-design:plus-outlined' }),
                ' ',
                $t('system.config.operation.createGroup'),
              ],
            ),
          ]);
        },
      };
    },
  },
  {
    fieldName: 'configType',
    label: $t('system.config.fields.configType'),
    component: 'RadioGroup',
    defaultValue: 'N',
    componentProps: {
      options: [
        { label: $t('system.config.config_type_options.N'), value: 'N' },
        { label: $t('system.config.config_type_options.Y'), value: 'Y' },
      ],
    },
  },
  {
    fieldName: 'status',
    label: $t('system.config.fields.status'),
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
    label: $t('system.config.fields.remark'),
    component: 'Textarea',
    componentProps: {
      placeholder: $t('system.config.form_placeholder.remark'),
      autoSize: { minRows: 2, maxRows: 4 },
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
});

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSubmit,
  onOpenChange(isOpen) {
    if (!isOpen) return;
    const data = modalApi.getData<SystemConfigApi.Config>();
    if (data) {
      formData.value = data;
      formApi.setValues(data);
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
    const data = (await formApi.getValues()) as Omit<
      SystemConfigApi.Config,
      'createdAt' | 'id' | 'updatedAt'
    >;
    if (data.group) data.group = data.group.trim();

    await (formData.value?.id
      ? updateConfigApi(formData.value.id, data)
      : createConfigApi(data));

    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const getModalTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('system.config.name')])
    : $t('ui.actionTitle.create', [$t('system.config.name')]),
);
</script>

<template>
  <Modal class="w-full max-w-[600px]" :title="getModalTitle">
    <Form :layout="isHorizontal ? 'horizontal' : 'vertical'" class="mx-4" />
  </Modal>
</template>
