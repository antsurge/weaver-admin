import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:layout-dashboard',
      order: 10,
      title: $t('page.adminuser.title'),
    },
    name: 'Adminuser',
    path: '/adminuser',
    children: [
      {
        name: 'Admin',
        path: '/admin',
        component: () => import('#/views/adminuser/admin/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.adminuser.admin'),
        },
      },
    ],
  },
];

export default routes;
