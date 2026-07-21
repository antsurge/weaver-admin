import { z } from '#/adapter/form';
import { $t } from '#/locales';

/**
 * 角色名称
 */
export const nameRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('organization.position.fields.name'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('organization.position.fields.name'), 30]),
  );

/**
 * 角色编码
 */
export const codeRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('organization.position.fields.code'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('organization.position.fields.code'), 30]),
  );
