import { $t } from '#/locales';
import { z } from '#/adapter/form';

const nameRule = z
  .string()
  .min(
    1,
    $t('ui.formRules.required', [
      $t('system.organization.fields.name'),
    ]),
  );

const parentIdRule = z
  .string()
// .min(
//   1,
//   $t('ui.formRules.required', [
//     $t('system.organization.fields.parentId'),
//   ]),
// );

const codeRule = z
  .string()
  .min(
    1,
    $t('ui.formRules.required', [
      $t('system.organization.fields.code'),
    ]),
  );

export {
  nameRule,
  parentIdRule,
  codeRule,
}
