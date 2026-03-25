import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:layout-dashboard',
      order: -1,
      title: $t('page.system.title'),
    },
    name: 'System',
    path: '/system',
    children: [
      {
        name: 'Dictionary',
        path: '/dictionary',
        component: () => import('#/views/system/dictionary/dict-type/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.system.dictionary'),
        },
      },
      {
        name: 'Admin',
        path: '/admin',
        component: () => import('#/views/system/admin/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.system.admin'),
        },
      },
    ],
  },
];

export default routes;
