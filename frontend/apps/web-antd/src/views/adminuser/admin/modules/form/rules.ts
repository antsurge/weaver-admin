import { z } from '#/adapter/form';
import { $t } from '#/locales';

/**
 * 真实姓名
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
 * 邮箱（可选，填写时校验格式）
 */
export const emailRule = z
  .string()
  .email($t('adminuser.admin.rules.emailInvalid'))
  .or(z.literal(''))
  .optional();

/**
 * 手机号（可选，填写时校验中国大陆手机号格式）
 */
export const phoneRule = z
  .string()
  .regex(/^1[3-9]\d{9}$/, $t('adminuser.admin.rules.phoneInvalid'))
  .or(z.literal(''))
  .optional();

/**
 * 密码（可选；填写时至少 6 位）
 */
export const passwordRule = z
  .string()
  .min(6, $t('adminuser.admin.rules.passwordTooShort', [6]))
  .or(z.literal(''))
  .optional();
