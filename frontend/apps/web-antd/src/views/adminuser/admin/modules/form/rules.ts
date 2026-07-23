import { z } from '#/adapter/form';
import { $t } from '#/locales';

/**
 * 职务名称
 */
export const realNameRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('adminuser.admin.fields.realName'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('adminuser.admin.fields.realName'), 30]),
  );

/**
 * 职务编码
 */
export const usernameRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('adminuser.admin.fields.username'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('adminuser.admin.fields.username'), 30]),
  );
