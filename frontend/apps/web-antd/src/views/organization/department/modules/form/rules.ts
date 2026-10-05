import { z } from '#/adapter/form';
import { $t } from '#/locales';

const nameRule = z
  .string()
  .min(
    1,
    $t('ui.formRules.required', [$t('organization.department.fields.name')]),
  );

const parentIDRule = z.string();
// .min(
//   1,
//   $t('ui.formRules.required', [
//     $t('organization.department.fields.parentID'),
//   ]),
// );

const typeRule = z
  .string()
  .min(
    1,
    $t('ui.formRules.selectRequired', [
      $t('organization.department.fields.type'),
    ]),
  );

export { nameRule, parentIDRule, typeRule };
