import { z } from '#/adapter/form';
import { $t } from '#/locales';

/**
 * 角色名称
 */
export const nameRule = z
  .string()
  .min(2, $t('ui.formRules.minLength', [$t('permission.role.fields.name'), 2]))
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('permission.role.fields.name'), 30]),
  );
