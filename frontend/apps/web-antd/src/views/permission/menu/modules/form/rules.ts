import { z } from '#/adapter/form';
import { $t } from '#/locales';

/**
 * 名称
 */
export const nameRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('permission.menu.fields.name'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.name'), 30]),
  );

/**
 * 标题
 */
export const titleRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('permission.menu.fields.title'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.title'), 30]),
  );  


/**
 * 路由路径
 */
export const pathRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('permission.menu.fields.path'), 2]),
  )
  .max(
    100,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.path'), 100]),
  );

/**
 * 组件路径
 */
export const componentRule = z.string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('permission.menu.fields.component'), 2]),
  )
  .max(
    100,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.component'), 100]),
  );

/**
 * 权限编码
 */
export const authCodeRule = z.string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('permission.menu.fields.authCode'), 2]),
  )
  .max(
    100,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.authCode'), 100]),
  );  

/**
 * 徽标内容
 */
export const badgeRule = z.string()
  .max(
    10,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.badge'), 10]),
  );  

/**
 * 链接路径
 */
export const linkUrlRule = z.string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('permission.menu.fields.linkUrl'), 2]),
  )
  .max(
    100,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.linkUrl'), 100]),
  );