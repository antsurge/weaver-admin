import type {
    ComponentRecordType,
    GenerateMenuAndRoutesOptions,
} from '@vben/types';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';
import { PermissionMenuApi } from "#/api/permission/menu"

import { message } from 'ant-design-vue';

import { getAllMenusApi } from '#/api';
import { BasicLayout, IFrameView } from '#/layouts';
import { $t } from '#/locales';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

/**
 * 页面组件映射
 *
 * key:
 * /src/views/system/user/index.vue
 *
 * value:
 * () => Promise<Component>
 */
const views = import.meta.glob(
    '/src/views/**/*.vue',
)

/**
 * 获取视图组件 - 支持多种路径格式
 *
 * 支持的格式：
 * - "permission/menu/index" → /src/views/permission/menu/index.vue
 * - "/permission/menu" → /src/views/permission/menu.vue (或 index.vue)
 * - "/menu" → /src/views/menu.vue (或 index.vue)
 */
function getViewComponent(component?: string) {
    if (!component) {
        console.warn('[Access] component 为空')
        return undefined
    }

    // 规范化路径：去掉前导斜杠和尾部斜杠
    let normalizedPath = component.replace(/^\/+|\/+$/g, '')

    if (!normalizedPath) {
        console.warn('[Access] component 路径为空:', component)
        return undefined
    }

    // 尝试多种路径组合
    const candidates = [
        `/src/views/${normalizedPath}.vue`,           // permission/menu/index.vue
        `/src/views/${normalizedPath}/index.vue`,     // permission/menu/index/
        `/src/views/${normalizedPath}.vue`,           // 如果原来没有 .vue 后缀
    ]

    for (const path of candidates) {
        const loader = views[path]
        if (loader) {
            console.log('[Access] ✅ 组件匹配成功:', path)
            return loader
        }
    }

    // 所有路径都未匹配
    console.error('[Access] ❌ 组件不存在, 尝试过的路径:', {
        original: component,
        normalized: normalizedPath,
        candidates,
        availablePaths: Object.keys(views).filter(p => p.includes(normalizedPath.split('/')[0])).slice(0, 5),
    })
    return undefined
}

function transformAccessRoutes(
    menus: PermissionMenuApi.PermissionMenu[],
    menuPaths: string[] = [],
): any[] {
    if (!Array.isArray(menus)) {
        console.error('菜单数据格式错误:', menus)
        return []
    }
    return menus
        .filter(menu => menu.type !== 'button' && menu.type !== 'menu_btn')
        .map(menu => {
            // 收集菜单页面路径（仅收集有实际页面的菜单）
            if ((menu.type === 'menu' || !menu.type) && menu.path) {
                menuPaths.push(menu.path)
            }

            // 判断是否为目录类型
            const isDirectory = menu.type === 'menu_dir' || menu.type === 'catalog'

            const route: any = {
                name: menu.code || menu.name,
                path: menu.path,
                meta: {
                    title: $t(menu.title),
                    icon: menu.icon,
                    order: menu.weight ?? 0,
                    badgeType: menu?.badgeType ?? "",
                    badge: menu?.badgeContent ?? "",
                    badgeVariants: menu?.badgeStyle ?? "",
                },
                // 🔧 修复1: 目录类型使用 BasicLayout 组件
                component: isDirectory
                    ? BasicLayout
                    : getViewComponent(menu.component),
            }

            // 递归处理子菜单
            if (menu.children?.length) {
                route.children = transformAccessRoutes(
                    menu.children,
                    menuPaths,
                )
            }

            // 调试日志
            if (isDirectory) {
                console.log(`[Access] 📁 目录路由: ${menu.name} (${menu.path})`)
            } else {
                const hasComponent = !!route.component
                console.log(`[Access] 📄 ${hasComponent ? '✅' : '❌'} 页面路由: ${menu.name} (${menu.path}) component=${menu.component}`)
            }

            return route
        })
}

async function generateAccess(options: GenerateMenuAndRoutesOptions) {
    const pageMap: ComponentRecordType = import.meta.glob('../views/**/*.vue');

    const layoutMap: ComponentRecordType = {
        BasicLayout,
        IFrameView,
    };

    return await generateAccessible(preferences.app.accessMode, {
        ...options,
        fetchMenuListAsync: async () => {
            message.loading({
                content: `${$t('common.loadingMenu')}...`,
                duration: 1.5,
            });
            return await getAllMenusApi();
        },
        // 可以指定没有权限跳转403页面
        forbiddenComponent,
        // 如果 route.meta.menuVisibleWithForbidden = true
        layoutMap,
        pageMap,
    });
}

export { transformAccessRoutes, generateAccess };