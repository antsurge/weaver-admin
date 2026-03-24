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

const parentIDRule = z
  .string()
// .min(
//   1,
//   $t('ui.formRules.required', [
//     $t('system.organization.fields.parentID'),
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
  parentIDRule,
  codeRule,
}
