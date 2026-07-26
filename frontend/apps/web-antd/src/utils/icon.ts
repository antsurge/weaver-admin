/**
 * 图标集工具函数
 * 支持 Iconify API 和本地 SVG 图标加载
 */

// 缓存已加载的图标列表
const ICONS_CACHE: Record<string, string[]> = {}

// 正在加载的 Promise（防止重复请求）
const LOADING_PROMISES: Record<string, Promise<string[]>> = {}

/**
 * 图标源配置接口
 */
export interface IconSourceConfig {
  /** 唯一标识 */
  key: string
  /** 显示名称 */
  label: string
  /** 图标集前缀 */
  prefix?: string
  /** 自定义加载函数 */
  loader?: () => Promise<string[]>
}

/**
 * 预置的图标源配置
 * 注意：prefix 必须是 Iconify API 支持的有效前缀
 */
export const PRESET_ICON_SOURCES: IconSourceConfig[] = [
  {
    key: 'ant-design',
    label: 'Ant Design',
    prefix: 'ant-design',
  },
  {
    key: 'bootstrap-icons',
    label: 'Bootstrap Icons',
    prefix: 'bi',  // Iconify 中 Bootstrap Icons 的正确 prefix
  },
  {
    key: 'lucide',
    label: 'Lucide',
    prefix: 'lucide',
  },
  {
    key: 'local-svg',
    label: '本地 SVG',
    prefix: 'svg',
  },
]

/**
 * 从 Iconify API 获取图标集数据
 */
async function fetchFromIconifyAPI(prefix: string): Promise<string[]> {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 10000)

  try {
    interface IconifyResponse {
      prefix: string
      total: number
      title: string
      uncategorized?: string[]
      categories?: Record<string, string[]>
    }

    const response: IconifyResponse = await fetch(
      `https://api.iconify.design/collection?prefix=${prefix}`,
      { signal: controller.signal }
    ).then((res) => res.json())

    clearTimeout(timeoutId)

    // 合并 uncategorized 和 categories
    const list = response.uncategorized || []
    if (response.categories) {
      for (const category in response.categories) {
        list.push(...(response.categories[category] || []))
      }
    }

    // 添加前缀
    return list.map((v) => `${prefix}:${v}`)
  } catch (error) {
    console.error(`[IconUtils] 从 API 加载 ${prefix} 图标集失败:`, error)
    return []
  } finally {
    clearTimeout(timeoutId)
  }
}

/**
 * 本地 SVG 图标模块缓存
 */
let localSVGModules: Record<string, string> | null = null

/**
 * 加载本地 SVG 图标
 * 从 src/assets/icons/ 目录动态导入所有 .svg 文件
 *
 * @returns 带前缀的图标名称数组，如 ['svg:logo', 'svg:dashboard']
 */
export async function loadLocalSVGs(): Promise<string[]> {
  // 检查缓存
  if (ICONS_CACHE['local-svg']?.length > 0) {
    return ICONS_CACHE['local-svg']
  }

  // 检查是否有正在进行的请求
  if (LOADING_PROMISES['local-svg']) {
    return LOADING_PROMISES['local-svg']
  }

  LOADING_PROMISES['local-svg'] = (async () => {
    try {
      // 使用 Vite 的 glob 导入功能
      const svgModules = import.meta.glob('#/assets/icons/**/*.svg', {
        eager: true,
        as: 'raw',
      })

      localSVGModules = svgModules as Record<string, string>
      const icons: string[] = []

      for (const path in svgModules) {
        // 从路径提取文件名作为图标名
        // 例如: /src/assets/icons/logo.svg -> logo
        const name = path.split('/').pop()?.replace('.svg', '') || ''
        const iconKey = `svg:${name}`

        icons.push(iconKey)

        // 动态注册到 Iconify（如果可用）
        try {
          const { addIcon } = await import('@iconify/vue')
          const svgContent = svgModules[path]

          // 提取 SVG 内容（去除 XML 声明等）
          let body = svgContent
          const match = svgContent.match(/<svg[^>]*>([\s\S]*)<\/svg>/i)
          if (match) {
            body = match[1]
          }

          addIcon(iconKey, {
            body: body.trim(),
            width: 24,
            height: 24,
          })
        } catch {
          // Iconify 注册失败时静默处理
          console.warn(`[IconUtils] 无法注册本地图标: ${iconKey}`)
        }
      }

      // 存入缓存
      ICONS_CACHE['local-svg'] = icons

      console.log(`[IconUtils] 已加载本地 SVG 图标，共 ${icons.length} 个`)

      return icons
    } catch (error) {
      console.error('[IconUtils] 加载本地 SVG 图标失败:', error)
      return []
    } finally {
      delete LOADING_PROMISES['local-svg']
    }
  })()

  return LOADING_PROMISES['local-svg']
}

/**
 * 获取指定前缀的图标名称列表
 *
 * @param prefix - 图标集前缀，如 'ant-design', 'lucide', 'mdi'
 * @returns 带前缀的图标名称数组，如 ['ant-design:user-outlined', 'lucide:home']
 *
 * @example
 * ```typescript
 * // 获取 ant-design 全部图标
 * const icons = await getIconList('ant-design')
 * // 返回: ['ant-design:account-book', 'ant-design:alert', ...]
 *
 * // 获取 lucide 全部图标
 * const lucideIcons = await getIconList('lucide')
 * // 返回: ['lucide:home', 'lucide:user', ...]
 * ```
 */
export async function getIconList(prefix: string): Promise<string[]> {
  // 本地 SVG 使用专门的加载函数
  if (prefix === 'local-svg' || prefix === 'svg') {
    return loadLocalSVGs()
  }

  // 1. 检查缓存
  if (ICONS_CACHE[prefix]?.length > 0) {
    return ICONS_CACHE[prefix]
  }

  // 2. 检查是否有正在进行的请求（防抖）
  if (LOADING_PROMISES[prefix]) {
    return LOADING_PROMISES[prefix]
  }

  // 3. 发起请求并缓存 Promise
  LOADING_PROMISES[prefix] = (async () => {
    try {
      // 使用 Iconify API 获取图标数据
      const iconList = await fetchFromIconifyAPI(prefix)

      if (iconList.length === 0) {
        console.warn(`[IconUtils] ${prefix} 图标集为空或加载失败`)
        return []
      }

      // 存入缓存
      ICONS_CACHE[prefix] = iconList

      console.log(`[IconUtils] 已加载 ${prefix} 图标集，共 ${iconList.length} 个图标`)

      return iconList
    } catch (error) {
      console.error(`[IconUtils] 加载 ${prefix} 图标集失败:`, error)
      return []
    } finally {
      // 清理 loading 状态
      delete LOADING_PROMISES[prefix]
    }
  })()

  return LOADING_PROMISES[prefix]
}

/**
 * 批量获取多个图标集的图标列表
 *
 * @param prefixes - 图标集前缀数组
 * @returns 合并后的图标名称数组
 *
 * @example
 * ```typescript
 * // 同时获取多个图标集
 * const allIcons = await getIconLists(['ant-design', 'lucide'])
 * ```
 */
export async function getIconLists(prefixes: string[]): Promise<string[]> {
  const results = await Promise.all(prefixes.map(getIconList))
  return results.flat()
}

/**
 * 清除指定前缀的图标缓存
 *
 * @param prefix - 图标集前缀，不传则清除所有缓存
 */
export function clearIconCache(prefix?: string) {
  if (prefix) {
    delete ICONS_CACHE[prefix]
  } else {
    Object.keys(ICONS_CACHE).forEach((key) => {
      delete ICONS_CACHE[key]
    })
  }
}
