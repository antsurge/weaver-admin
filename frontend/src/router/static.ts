import type { RouteRecordRaw } from "vue-router";
import {adminBaseRoutePath} from "@/router/static/adminBase"

const pageTitle = (name: string): string => {
  return `pagesTitle.${name}`;
};

/*
 * 静态路由
 * 自动加载 ./static 目录的所有文件，并 push 到以下数组
 */
const staticRoutes:Array<RouteRecordRaw> = [
    {
        // 首页
        path: `${adminBaseRoutePath}/login`,
        name: 'adminLogin',
        component: () => import('@/views/login.vue'),
        meta: {
            title: pageTitle('adminLogin'),
        },
    },
    {
        path: '/:path(.*)*',
        redirect: '/404',
    },
    {
        // 404
        path: '/404',
        name: 'notFound',
        component: () => import('@/views/common/error/404.vue'),
        meta: {
            title: pageTitle('notFound'), // 页面不存在
        },
    },
    {
        // 无权限访问
        path: '/401',
        name: 'noPower',
        component: () => import('@/views/common/error/401.vue'),
        meta: {
            title: pageTitle('noPower'),
        },
    },
]

const staticFiles: Record<string, { default?: RouteRecordRaw }> = import.meta.glob('./static/*.ts', { eager: true })
for (const key in staticFiles) {
    const mod = staticFiles[key]
    if (mod && mod.default) staticRoutes.push(mod.default)
}

export default staticRoutes