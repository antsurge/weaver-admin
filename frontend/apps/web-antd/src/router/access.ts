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

import {
    PermissionTypeOptionsValueIframe,
    PermissionTypeOptionsValueLink,
    PermissionTypeOptionsValueCatalog,
    PermissionTypeOptionsValueAction,
    PermissionTypeOptionsValueMenu
} from '#/views/permission/menu/data';

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

/**
 * 为特殊类型菜单（外链 / 内嵌）构造占位路由
 *
 * 为什么不直接用后端 linkUrl 作为 path？
 * - 外链/iframe 的 URL 是绝对地址（https://...），直接塞进 vue-router 会污染
 *   useRoute().path、侧边栏高亮、面包屑、Tab 栏等所有依赖路由的逻辑。
 * - 用占位 path + meta 标识字段（meta.link / meta.iframeSrc）能保持 Vben 框架约定，
 *   use-navigation / IFrameRouterView 各自基于 meta 字段触发对应行为。
 *
 * @param menu 后端返回的菜单项
 * @param type 特殊类型：'link' | 'iframe'
 * @returns 占位路由配置；非特殊类型或缺 linkUrl 时返回 null
 */
type SpecialMenuType = 'link' | 'iframe'

interface SpecialRouteOverride {
    name: string
    path: string
    meta: Record<string, unknown>
    component: any
}

function buildSpecialRoute(
    menu: PermissionMenuApi.PermissionMenu,
    type: SpecialMenuType,
): SpecialRouteOverride | null {
    const linkUrl = menu.linkUrl
    if (!linkUrl) {
        return null
    }

    const code = menu.code || menu.name || String(menu.id ?? '')

    if (type === PermissionTypeOptionsValueLink) {
        return {
            name: `Link_${code}`,
            path: `/link/${code}`,
            meta: {
                link: linkUrl,
                openInNewWindow: true,
            },
            component: undefined,
        }
    }

    // iframe
    return {
        name: `Iframe_${code}`,
        path: `/iframe/${code}`,
        meta: {
            iframeSrc: linkUrl,
        },
        component: IFrameView,
    }
}

function transformAccessRoutes(
  menus: PermissionMenuApi.PermissionMenu[],
  menuPaths: string[] = [],
  accessCodes: string[] = [],
): any[] {
  if (!Array.isArray(menus)) {
    console.error('菜单数据格式错误:', menus)
    return []
  }

  const routes: any[] = []

  menus.forEach(menu => {

    // 所有类型，只要存在 authCode 都收集
    if (menu.authCode) {
      if (!accessCodes.includes(menu.authCode)) {
        accessCodes.push(menu.authCode)
      }
    }

    // 按钮类型不生成路由
    if (menu.type === PermissionTypeOptionsValueAction) {
      return
    }


    // 收集菜单页面路径
    if (
      (menu.type === PermissionTypeOptionsValueMenu || !menu.type)
      && menu.path
    ) {
      menuPaths.push(menu.path)
    }


    const isDirectory =
      menu.type === PermissionTypeOptionsValueCatalog

    const isLink =
      menu.type === PermissionTypeOptionsValueLink

    const isIframe =
      menu.type === PermissionTypeOptionsValueIframe


    const route: any = {
      name: menu.code || menu.name,
      path: menu.path,
      meta: {
        title: $t(menu.title),
        icon: menu.icon,
        order: menu.weight ?? 0,
        badgeType: menu.badgeType ?? '',
        badge: menu.badge ?? '',
        badgeVariants: menu.badgeVariants ?? '',
      },

      component: isDirectory
        ? BasicLayout
        : getViewComponent(menu.component),
    }


    if (isLink || isIframe) {
      const override = buildSpecialRoute(
        menu,
        isLink ? 'link' : 'iframe',
      )

      if (override) {
        route.name = override.name
        route.path = override.path
        route.component = override.component
        Object.assign(route.meta, override.meta)
      }
    }


    if (menu.children?.length) {
      route.children = transformAccessRoutes(
        menu.children,
        menuPaths,
        accessCodes,
      )
    }


    routes.push(route)
  })


  return routes
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