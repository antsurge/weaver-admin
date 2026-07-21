import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:layout-dashboard',
      order: -1,
      title: $t('page.organization.title'),
    },
    name: 'Organization',
    path: '/organization',
    children: [
      {
        name: 'Department',
        path: '/department',
        component: () => import('#/views/organization/department/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.organization.department'),
        },
      },
      {
        name: 'Position',
        path: '/position',
        component: () => import('#/views/organization/position/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.organization.position'),
        },
      },
    ],
  },
];

export default routes;
