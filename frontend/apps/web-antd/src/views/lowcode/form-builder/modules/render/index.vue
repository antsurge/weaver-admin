<script lang="ts" setup>
import type { FormSchemaApi } from '#/api/lowcode/form-schema';

import { computed, reactive, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  Button,
  CheckboxGroup,
  DatePicker,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  RadioGroup,
  Select,
  Switch,
  Textarea,
} from 'ant-design-vue';

import { submitFormDataApi } from '#/api/lowcode/form-schema';
import { $t } from '#/locales';

/** 运行时字段类型（与设计器一致） */
type FieldType =
  | 'checkbox'
  | 'date'
  | 'input'
  | 'number'
  | 'radio'
  | 'select'
  | 'switch'
  | 'textarea';

interface RenderField {
  key: string;
  label: string;
  type: FieldType;
  options?: string[];
}

interface SchemaJson {
  version?: number;
  fields: RenderField[];
}

const emit = defineEmits<{
  success: [];
}>();

const form = ref<FormSchemaApi.FormSchema>();
const fields = ref<RenderField[]>([]);
const values = reactive<Record<string, any>>({});
const submitting = ref(false);

const [Modal, modalApi] = useVbenModal({
  footer: false,
  onOpenChange(isOpen) {
    if (!isOpen) return;
    const data = modalApi.getData<FormSchemaApi.FormSchema>();
    if (!data) return;
    form.value = data;
    const schemaJson = data.schemaJson as SchemaJson | undefined;
    fields.value = schemaJson?.fields ?? [];
    // 重置提交数据
    Object.keys(values).forEach((k) => delete values[k]);
    fields.value.forEach((f) => {
      if (f.type === 'switch') {
        values[f.key] = false;
      } else if (f.type === 'checkbox') {
        values[f.key] = [];
      } else {
        values[f.key] = undefined;
      }
    });
  },
});

const canSubmit = computed(() => fields.value.length > 0 && !!form.value?.code);

function fieldOptions(field: RenderField) {
  return (field.options || []).map((o) => ({ label: o, value: o }));
}

async function onSubmit() {
  if (!form.value?.code) return;
  submitting.value = true;
  try {
    const data: Record<string, unknown> = {};
    fields.value.forEach((f) => {
      data[f.key] = values[f.key];
    });
    await submitFormDataApi(form.value.code, data);
    message.success($t('lowcode.formBuilder.render.submitSuccess'));
    emit('success');
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <Modal
    class="w-full max-w-[720px]"
    :title="form?.name || $t('lowcode.formBuilder.render.title')"
  >
    <Form layout="vertical">
      <FormItem
        v-for="field in fields"
        :key="field.key"
        :label="field.label"
        class="mb-4"
      >
        <template v-if="field.type === 'textarea'">
          <Textarea
            v-model:value="values[field.key]"
            :rows="3"
            :placeholder="field.label"
          />
        </template>
        <template v-else-if="field.type === 'select'">
          <Select
            v-model:value="values[field.key]"
            allow-clear
            :placeholder="field.label"
            :options="fieldOptions(field)"
          />
        </template>
        <template v-else-if="field.type === 'radio'">
          <RadioGroup
            v-model:value="values[field.key]"
            :options="fieldOptions(field)"
          />
        </template>
        <template v-else-if="field.type === 'checkbox'">
          <CheckboxGroup
            v-model:value="values[field.key]"
            :options="fieldOptions(field)"
          />
        </template>
        <template v-else-if="field.type === 'date'">
          <DatePicker
            v-model:value="values[field.key]"
            value-format="YYYY-MM-DD"
            style="width: 100%"
            :placeholder="field.label"
          />
        </template>
        <template v-else-if="field.type === 'number'">
          <InputNumber
            v-model:value="values[field.key]"
            style="width: 100%"
            :placeholder="field.label"
          />
        </template>
        <template v-else-if="field.type === 'switch'">
          <Switch v-model:checked="values[field.key]" />
        </template>
        <template v-else>
          <Input
            v-model:value="values[field.key]"
            allow-clear
            :placeholder="field.label"
          />
        </template>
      </FormItem>

      <FormItem v-if="fields.length === 0">
        <div class="py-8 text-center text-gray-400">
          {{ $t('lowcode.formBuilder.designer.noField') }}
        </div>
      </FormItem>

      <FormItem>
        <div class="flex justify-end gap-2">
          <Button :disabled="submitting" @click="modalApi.close()">
            {{ $t('lowcode.formBuilder.designer.cancel') }}
          </Button>
          <Button
            type="primary"
            :disabled="!canSubmit"
            :loading="submitting"
            @click="onSubmit"
          >
            {{ $t('lowcode.formBuilder.render.submit') }}
          </Button>
        </div>
      </FormItem>
    </Form>
  </Modal>
</template>
