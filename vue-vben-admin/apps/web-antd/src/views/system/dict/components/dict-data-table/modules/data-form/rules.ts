import { z } from '#/adapter/form';
import { $t } from '#/locales';

export const dictDataFormRules = {
  dictTypeId: z.string({
    required_error: $t('ui.formRules.required', [
      $t('system.dict_data.fields.dictTypeID'),
    ]),
  }),
  label: z
    .string()
    .min(1, $t('ui.formRules.minLength', [$t('system.dict_data.fields.label'), 1]))
    .max(
      50,
      $t('ui.formRules.maxLength', [$t('system.dict_data.fields.label'), 50]),
    ),
  value: z
    .string()
    .min(1, $t('ui.formRules.minLength', [$t('system.dict_data.fields.value'), 1]))
    .max(
      100,
      $t('ui.formRules.maxLength', [$t('system.dict_data.fields.value'), 100]),
    ),
} as const;

