import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:layout-dashboard',
      order: -1,
      title: $t('page.permission.title'),
    },
    name: 'Permission',
    path: '/permission',
    children: [
      {
        name: 'Menu',
        path: '/menu',
        component: () => import('#/views/permission/menu/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.permission.menu'),
        },
      },
      {
        name: 'Role',
        path: '/role',
        component: () => import('#/views/permission/role/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.permission.role'),
        },
      },
    ],
  },
];

export default routes;
