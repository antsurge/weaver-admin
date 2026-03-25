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
        name: 'Admin',
        path: '/admin',
        component: () => import('#/views/system/admin/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.system.admin'),
        },
      },
      {
        name: 'Role',
        path: '/role',
        component: () => import('#/views/system/role/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.system.role'),
        },
      },
      {
        name: 'Permission',
        path: '/permission',
        component: () => import('#/views/system/permission/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.system.permission'),
        },
      },
      {
        name: 'Dict',
        path: '/dict',
        component: () => import('#/views/system/dict/index.vue'),
        meta: {
          icon: 'lucide:area-chart',
          title: $t('page.system.dict'),
        },
      },
    ],
  },
];

export default routes;
