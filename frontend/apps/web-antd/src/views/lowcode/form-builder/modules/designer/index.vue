<script lang="ts" setup>
import type { FormSchemaApi } from '#/api/lowcode/form-schema';

import { computed, h, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  Button,
  Checkbox,
  Empty,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  Radio,
  Select,
  Switch,
  Tag,
  Textarea,
} from 'ant-design-vue';

import { updateFormSchemaApi } from '#/api/lowcode/form-schema';
import { $t } from '#/locales';

const emit = defineEmits<{
  success: [];
}>();

/** 字段组件类型 */
type FieldType =
  | 'checkbox'
  | 'date'
  | 'input'
  | 'number'
  | 'radio'
  | 'select'
  | 'switch'
  | 'textarea';

interface FieldDef {
  key: string;
  label: string;
  type: FieldType;
  options?: string[];
}

const formData = ref<FormSchemaApi.FormSchema>();
const fieldKeySeq = ref(0);

/** 已添加字段 */
const fields = ref<FieldDef[]>([]);
/** 当前选中字段 key */
const activeKey = ref('');

const COMPONENT_LIST = computed(() => [
  { label: $t('lowcode.formBuilder.designer.input'), type: 'input' },
  { label: $t('lowcode.formBuilder.designer.textarea'), type: 'textarea' },
  { label: $t('lowcode.formBuilder.designer.select'), type: 'select' },
  { label: $t('lowcode.formBuilder.designer.radio'), type: 'radio' },
  { label: $t('lowcode.formBuilder.designer.checkbox'), type: 'checkbox' },
  { label: $t('lowcode.formBuilder.designer.date'), type: 'date' },
  { label: $t('lowcode.formBuilder.designer.number'), type: 'number' },
  { label: $t('lowcode.formBuilder.designer.switch'), type: 'switch' },
]);

const activeField = computed(() =>
  fields.value.find((f) => f.key === activeKey.value),
);

const selectedComponent = ref('input');
const newFieldLabel = ref('');

/** 添加字段到画布 */
function addField() {
  const type = selectedComponent.value as FieldType;
  const seq = ++fieldKeySeq.value;
  const label =
    newFieldLabel.value || `${$t('lowcode.formBuilder.fields.name')} ${seq}`;
  const field: FieldDef = {
    key: `f_${Date.now()}_${seq}`,
    label,
    type,
    options:
      type === 'select' || type === 'radio' || type === 'checkbox'
        ? ['选项1', '选项2', '选项3']
        : undefined,
  };
  fields.value.push(field);
  activeKey.value = field.key;
  newFieldLabel.value = '';
}

/** 删除字段 */
function removeField(key: string) {
  fields.value = fields.value.filter((f) => f.key !== key);
  if (activeKey.value === key) activeKey.value = '';
}

/** 更新当前选中字段的选项（textarea 行分隔编辑） */
function updateActiveFieldOptions(e: Event) {
  const field = activeField.value;
  if (!field) return;
  const value = (e.target as HTMLInputElement).value || '';
  field.options = value ? value.split('\n').filter(Boolean) : [];
}

/** 调整字段顺序 */
function moveField(key: string, dir: -1 | 1) {
  const idx = fields.value.findIndex((f) => f.key === key);
  const target = idx + dir;
  if (idx < 0 || target < 0 || target >= fields.value.length) return;
  const arr = [...fields.value];
  const a = arr[idx]!;
  const b = arr[target]!;
  arr[idx] = b;
  arr[target] = a;
  fields.value = arr;
}

/** 渲染字段对应组件预览 */
function renderControl(field: FieldDef) {
  const opts = (field.options || []).map((o) => ({ label: o, value: o }));
  switch (field.type) {
    case 'checkbox': {
      return h(Checkbox.Group, {}, () =>
        (field.options || []).map((o) =>
          h(Checkbox, { value: o, key: o }, { default: () => o }),
        ),
      );
    }
    case 'date': {
      return h(Input, { placeholder: field.label });
    }
    case 'number': {
      return h(InputNumber, { style: 'width: 100%', placeholder: field.label });
    }
    case 'radio': {
      return h(Radio.Group, {}, () =>
        (field.options || []).map((o) =>
          h(Radio, { value: o, key: o }, { default: () => o }),
        ),
      );
    }
    case 'select': {
      return h(Select, { placeholder: field.label, options: opts });
    }
    case 'switch': {
      return h(Switch);
    }
    case 'textarea': {
      return h(Textarea, { rows: 2, placeholder: field.label });
    }
    default: {
      return h(Input, { placeholder: field.label });
    }
  }
}

/** 保存表单（schemaJson 序列化字段定义） */
async function onSave() {
  if (!formData.value) return;
  modalApi.lock();
  try {
    const schemaJson = {
      version: 1,
      fields: fields.value.map((f) => ({
        key: f.key,
        label: f.label,
        type: f.type,
        options: f.options,
      })),
    };
    const data = {
      name: formData.value.name,
      code: formData.value.code,
      description: formData.value.description,
      status: formData.value.status,
      schemaJson,
    };
    await updateFormSchemaApi(formData.value.id, data);
    message.success($t('lowcode.formBuilder.designer.saveSuccess'));
    modalApi.close();
    emit('success');
  } finally {
    modalApi.unlock();
  }
}

const [Modal, modalApi] = useVbenModal({
  onConfirm: onSave,
  onOpenChange(isOpen) {
    if (!isOpen) return;
    const data = modalApi.getData<FormSchemaApi.FormSchema>();
    if (!data) return;
    formData.value = data;
    const savedFields = (data.schemaJson?.fields || []) as any[];
    fields.value = savedFields.map((f: any, idx: number) => ({
      key: f.key || `f_${idx}`,
      label: f.label || f.key || '',
      type: (f.type || 'input') as FieldType,
      options: f.options,
    }));
    fieldKeySeq.value = fields.value.length;
    activeKey.value = fields.value[0]?.key || '';
  },
});
</script>

<template>
  <Modal
    class="w-full max-w-[1000px]"
    :title="$t('lowcode.formBuilder.designer.title')"
    :footer="false"
    :mask-closable="false"
  >
    <div class="flex h-[60vh] gap-4 p-2">
      <!-- 左侧：组件面板 -->
      <div
        class="flex w-[200px] shrink-0 flex-col gap-2 overflow-auto rounded border p-3"
      >
        <div class="mb-1 text-sm font-medium">
          {{ $t('lowcode.formBuilder.designer.addField') }}
        </div>
        <div v-for="c in COMPONENT_LIST" :key="c.type">
          <Button
            class="w-full justify-start"
            @click="selectedComponent = c.type"
            :type="selectedComponent === c.type ? 'primary' : 'default'"
          >
            {{ c.label }}
          </Button>
        </div>
        <Input
          v-model:value="newFieldLabel"
          :placeholder="$t('lowcode.formBuilder.designer.fieldLabel')"
          class="mt-2"
        />
        <Button type="primary" block @click="addField">
          + {{ $t('lowcode.formBuilder.designer.addField') }}
        </Button>
      </div>

      <!-- 中间：画布 -->
      <div class="flex-1 overflow-auto rounded border p-3">
        <div
          v-if="fields.length === 0"
          class="flex h-full items-center justify-center"
        >
          <Empty :description="$t('lowcode.formBuilder.designer.noField')" />
        </div>
        <div
          v-for="field in fields"
          :key="field.key"
          class="mb-2 cursor-pointer rounded border p-2 transition"
          :class="activeKey === field.key ? 'border-primary bg-primary/5' : ''"
          @click="activeKey = field.key"
        >
          <div class="mb-1 flex items-center justify-between">
            <Tag :color="activeKey === field.key ? 'blue' : 'default'">
              {{ field.label }}
            </Tag>
            <div class="flex gap-1">
              <Button size="small" @click.stop="moveField(field.key, -1)">
                ↑
              </Button>
              <Button size="small" @click.stop="moveField(field.key, 1)">
                ↓
              </Button>
              <Button size="small" danger @click.stop="removeField(field.key)">
                ✕
              </Button>
            </div>
          </div>
          <component :is="renderControl(field)" />
        </div>
      </div>

      <!-- 右侧：属性配置 -->
      <div class="w-[260px] shrink-0 overflow-auto rounded border p-3">
        <div class="mb-3 text-sm font-medium">
          {{ $t('lowcode.formBuilder.designer.fieldName') }}
        </div>
        <template v-if="activeField">
          <Form layout="vertical" size="small">
            <FormItem :label="$t('lowcode.formBuilder.designer.fieldLabel')">
              <Input v-model:value="activeField.label" />
            </FormItem>
            <FormItem :label="$t('lowcode.formBuilder.designer.component')">
              <Select
                v-model:value="activeField.type"
                :options="
                  COMPONENT_LIST.map((c) => ({ label: c.label, value: c.type }))
                "
              />
            </FormItem>
            <FormItem
              v-if="
                activeField.type === 'select' ||
                activeField.type === 'radio' ||
                activeField.type === 'checkbox'
              "
              :label="`${$t('lowcode.formBuilder.designer.component')} Options`"
            >
              <Input
                :value="activeField.options?.join('\n')"
                @change="updateActiveFieldOptions"
              />
            </FormItem>
          </Form>
        </template>
        <div v-else class="text-sm text-gray-400">
          {{ $t('lowcode.formBuilder.designer.noField') }}
        </div>
      </div>
    </div>

    <!-- 底部操作 -->
    <div class="mt-3 flex justify-end gap-2 pr-2">
      <Button @click="modalApi.close()">
        {{ $t('lowcode.formBuilder.designer.cancel') }}
      </Button>
      <Button type="primary" @click="onSave">
        {{ $t('lowcode.formBuilder.designer.save') }}
      </Button>
    </div>
  </Modal>
</template>
