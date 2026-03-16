import { z } from '#/adapter/form';
import { $t } from '#/locales';

/**
 * 岗位名称
 */
export const nameRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('system.position.fields.name'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('system.position.fields.name'), 30]),
  );

/**
 * 岗位编码
 */
export const codeRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('system.position.fields.code'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('system.position.fields.code'), 30]),
  );
