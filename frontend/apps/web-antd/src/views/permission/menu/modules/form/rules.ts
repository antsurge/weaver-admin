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
 * 权限编码
 */
export const codeRule = z
  .string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('permission.menu.fields.code'), 2]),
  )
  .max(
    30,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.code'), 30]),
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
    30,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.path'), 30]),
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

// rules: z
//   .string()
//   .min(2, $t('ui.formRules.minLength', [$t('permission.menu.fields.code'), 2]))
//   .max(30, $t('ui.formRules.maxLength', [$t('permission.menu.fields.code'), 30])),
// .refine(
//   async (value: string) =>
//     !(await isPermissionNameExists(value, formData.value?.id)),
//   (value) => ({
//     message: $t('ui.formRules.alreadyExists', [
//       $t('system.menu.menuName'),
//       value,
//     ]),
//   }),
// ),
/**
 * 组件路径
 */
export const urlRule = z.string()
  .min(
    2,
    $t('ui.formRules.minLength', [$t('permission.menu.fields.url'), 2]),
  )
  .max(
    100,
    $t('ui.formRules.maxLength', [$t('permission.menu.fields.url'), 100]),
  );
