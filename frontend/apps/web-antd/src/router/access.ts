import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
} from '@vben/types';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';
import { convertRoutes, normalizeViewPath } from '@vben/utils';

import { message } from 'ant-design-vue';

import { PermissionMenuApi } from '#/api/permission/menu';
import { BasicLayout, IFrameView } from '#/layouts';
import { $t } from '#/locales';
import {
  PermissionTypeOptionsValueAction,
  PermissionTypeOptionsValueCatalog,
  PermissionTypeOptionsValueIframe,
  PermissionTypeOptionsValueLink,
  PermissionTypeOptionsValueMenu,
} from '#/views/permission/menu/data';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

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
type SpecialMenuType = 'iframe' | 'link';

interface SpecialRouteOverride {
  name: string;
  path: string;
  meta: Record<string, unknown>;
  component: any;
}

function buildSpecialRoute(
  menu: PermissionMenuApi.PermissionMenu,
  type: SpecialMenuType,
): null | SpecialRouteOverride {
  const linkUrl = menu.linkUrl;
  if (!linkUrl) {
    return null;
  }

  const code = menu.code || menu.name || String(menu.id ?? '');

  if (type === PermissionTypeOptionsValueLink) {
    return {
      name: `Link_${code}`,
      path: `/link/${code}`,
      meta: {
        link: linkUrl,
        openInNewWindow: true,
      },
      component: undefined,
    };
  }

  // iframe（组件用字符串标识，由 layoutMap 解析）
  return {
    name: `Iframe_${code}`,
    path: `/iframe/${code}`,
    meta: {
      iframeSrc: linkUrl,
    },
    component: 'IFrameView',
  };
}

/**
 * 找到 menuTree 中第一个可访问的 type=menu 页面路径
 *
 * 收集规则与 transformAccessRoutes 中的 accessMenuPaths 保持一致：
 * - 仅 type === 'menu'（或未声明 type）且有 path 的节点算一个可访问页面
 * - 目录（catalog）/ 按钮（action）/ 外链（link）/ 内嵌（iframe）不计入
 * - 先序遍历：父级 menu 优先于其子级
 *
 * @returns 第一个可访问菜单路径；不存在则返回空串
 */
function findFirstMenuPath(menus: PermissionMenuApi.PermissionMenu[]): string {
  if (!Array.isArray(menus)) {
    return '';
  }

  for (const menu of menus) {
    if (
      (menu.type === PermissionTypeOptionsValueMenu || !menu.type) &&
      menu.path
    ) {
      return menu.path;
    }

    if (menu.children?.length) {
      const childPath = findFirstMenuPath(menu.children);
      if (childPath) {
        return childPath;
      }
    }
  }

  return '';
}

function transformAccessRoutes(
  menus: PermissionMenuApi.PermissionMenu[],
  menuPaths: string[] = [],
  accessCodes: string[] = [],
): any[] {
  if (!Array.isArray(menus)) {
    console.error('菜单数据格式错误:', menus);
    return [];
  }

  const routes: any[] = [];

  menus.forEach((menu) => {
    // 所有类型，只要存在 authCode 都收集
    if (menu.authCode && !accessCodes.includes(menu.authCode)) {
      accessCodes.push(menu.authCode);
    }

    // 按钮类型不生成路由
    if (menu.type === PermissionTypeOptionsValueAction) {
      return;
    }

    // 收集菜单页面路径
    if (
      (menu.type === PermissionTypeOptionsValueMenu || !menu.type) &&
      menu.path
    ) {
      menuPaths.push(menu.path);
    }

    const isDirectory = menu.type === PermissionTypeOptionsValueCatalog;

    const isLink = menu.type === PermissionTypeOptionsValueLink;

    const isIframe = menu.type === PermissionTypeOptionsValueIframe;

    const route: any = {
      name: menu.code || menu.name,
      path: menu.path,
      meta: {
        title: $t(menu.title ?? ''),
        icon: menu.icon,
        order: menu.weight ?? 0,
        badgeType: menu.badgeType ?? '',
        badge: menu.badge ?? '',
        badgeVariants: menu.badgeVariants ?? '',
      },

      // 组件使用字符串标识，由 generateRoutesByBackend 通过 layoutMap/pageMap 解析
      // catalog → BasicLayout；menu 类型 → 后端存储的组件路径
      component: isDirectory ? 'BasicLayout' : menu.component,
    };

    if (isLink || isIframe) {
      const override = buildSpecialRoute(menu, isLink ? 'link' : 'iframe');

      if (override) {
        route.name = override.name;
        route.path = override.path;
        route.component = override.component;
        Object.assign(route.meta, override.meta);
      }
    }

    if (menu.children?.length) {
      route.children = transformAccessRoutes(
        menu.children,
        menuPaths,
        accessCodes,
      );
    }

    routes.push(route);
  });

  return routes;
}

async function generateAccess(options: GenerateMenuAndRoutesOptions) {
  const pageMap: ComponentRecordType = import.meta.glob('../views/**/*.vue');

  const layoutMap: ComponentRecordType = {
    BasicLayout,
    IFrameView,
  };

  const normalizePageMap: ComponentRecordType = {};
  for (const [key, value] of Object.entries(pageMap)) {
    normalizePageMap[normalizeViewPath(key)] = value;
  }

  // transformAccessRoutes 产出的组件是后端字符串（catalog → 'BasicLayout'，
  // menu → 'admin/user/index'），只有 backend 模式的 generateRoutesByBackend
  // 会将其解析为真实组件。若 preferences.accessMode 被 localStorage 缓存覆盖为
  // frontend/mixed，字符串组件会原样进入路由表，导航时触发 vue-router 的
  // "Invalid route component"(R0027)。因此在这里统一提前解析，与 accessMode 解耦。
  const resolvedRoutes = convertRoutes(
    (options.routes ?? []) as any,
    layoutMap,
    normalizePageMap,
  );

  return await generateAccessible(preferences.app.accessMode, {
    ...options,
    // 传入已解析的路由，保证任何模式下路由表里都不存在字符串组件
    routes: resolvedRoutes as any,
    // 路由已在 guard 中根据 current-user 的 menuTree 构建完成（transformAccessRoutes），
    // 直接复用，不再请求后端不存在的 /menu/all
    fetchMenuListAsync: async () => {
      message.loading({
        content: `${$t('common.loadingMenu')}...`,
        duration: 1.5,
      });
      return resolvedRoutes as any;
    },
    // 可以指定没有权限跳转403页面
    forbiddenComponent,
    // 如果 route.meta.menuVisibleWithForbidden = true
    layoutMap,
    pageMap,
  });
}

export { findFirstMenuPath, generateAccess, transformAccessRoutes };
