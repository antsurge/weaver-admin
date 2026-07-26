<script setup lang="ts">
import { computed, ref, useAttrs, watch, watchEffect } from 'vue'

import { usePagination } from '@vben/hooks'
import {
  EmptyIcon,
  Grip,
  listIcons,
} from '@vben/icons'
import { $t } from '@vben/locales'

// 使用 ant-design-vue 组件
import {
  Button,
  Input,
  Popover,
  Tabs,
  TabPane,
  Tooltip,
  Pagination,
} from 'ant-design-vue'

// iconify 图标组件
import { Icon as IconifyIcon } from '@iconify/vue'

import { objectOmit, refDebounced } from '@vueuse/core'

import type { IconSourceConfig } from '#/utils/icon'
import {
  getIconList,
  loadLocalSVGs,
  PRESET_ICON_SOURCES,
} from '#/utils/icon'

interface Props {
  pageSize?: number
  /** 图标集的名字 */
  prefix?: string
  /** 是否自动请求API以获得图标集的数据.提供prefix时有效 */
  autoFetchApi?: boolean
  /**
   * 图标列表
   */
  icons?: string[]
  /** Input组件 */
  inputComponent?: any
  /** 图标插槽名，预览图标将被渲染到此插槽中 */
  iconSlot?: string
  /** input组件的值属性名称 */
  modelValueProp?: string
  /** 图标样式 */
  iconClass?: string
  type?: 'icon' | 'input'
  /** 启用的图标源列表 */
  sources?: string[]
  /** 自定义图标源配置 */
  customSources?: IconSourceConfig[]
}

const props = withDefaults(defineProps<Props>(), {
  prefix: 'ant-design',
  pageSize: 36,
  icons: () => [],
  iconSlot: 'default',
  iconClass: 'size-4',
  autoFetchApi: true,
  modelValueProp: 'modelValue',
  inputComponent: undefined,
  type: 'input',
  sources: () => ['ant-design', 'bootstrap-icons', 'lucide', 'local-svg'],
})

const emit = defineEmits<{
  change: [string]
}>()

const attrs = useAttrs()

const modelValue = defineModel({ default: '', type: String })

const visible = ref(false)
const currentSelect = ref('')
const keyword = ref('')
const keywordDebounce = refDebounced(keyword, 300)
const innerIcons = ref<string[]>([])

// ====== 多图标源支持 ======
const activeTab = ref('ant-design')
const sourceIconsMap = ref<Record<string, string[]>>({})
const sourceLoadingMap = ref<Record<string, boolean>>({})

// 获取可用的图标源列表
const availableSources = computed(() => {
  const sources = [...PRESET_ICON_SOURCES]

  // 添加自定义图标源
  if (props.customSources?.length) {
    sources.push(...props.customSources)
  }

  // 过滤启用的图标源
  return sources.filter((s) => props.sources.includes(s.key))
})

// 切换 Tab 时加载对应图标源
async function handleTabChange(tabKey: string) {
  activeTab.value = tabKey

  // 如果该源的图标已加载，直接返回
  if (sourceIconsMap.value[tabKey]?.length > 0) {
    return
  }

  // 标记加载状态
  sourceLoadingMap.value[tabKey] = true

  try {
    let icons: string[] = []

    // 查找对应图标源的配置
    const sourceConfig = availableSources.value.find((s) => s.key === tabKey)

    switch (tabKey) {
      case 'ant-design':
      case 'bootstrap-icons':
      case 'lucide':
        // 使用配置中的 prefix 调用 API，而不是 key
        icons = await getIconList(sourceConfig?.prefix || tabKey)
        break
      case 'local-svg':
        icons = await loadLocalSVGs()
        break
      default:
        // 自定义图标源
        if (sourceConfig?.loader) {
          icons = await sourceConfig.loader()
        }
    }

    sourceIconsMap.value[tabKey] = icons
  } catch (error) {
    console.error(`[EnhancedIconPicker] 加载 ${tabKey} 图标源失败:`, error)
    sourceIconsMap.value[tabKey] = []
  } finally {
    sourceLoadingMap.value[tabKey] = false
  }
}

// 当前激活的图标列表
const currentList = computed(() => {
  const tabKey = activeTab.value
  const icons = sourceIconsMap.value[tabKey] || []

  // 如果有自定义 icons 且当前是默认 tab
  if (
    props.prefix &&
    props.autoFetchApi &&
    props.icons.length > 0 &&
    !sourceIconsMap.value[tabKey]
  ) {
    const localIcons = listIcons('', props.prefix)
    if (localIcons.length > 0) return localIcons
    return props.icons
  }

  return icons
})

const showList = computed(() => {
  return currentList.value.filter((item) =>
    item.includes(keywordDebounce.value),
  )
})

const { paginationList, total, setCurrentPage, currentPage } =
  usePagination(showList, props.pageSize)

watchEffect(() => {
  currentSelect.value = modelValue.value
})

watch(
  () => currentSelect.value,
  (v) => {
    emit('change', v)
  },
)

const handleClick = (icon: string) => {
  currentSelect.value = icon
  modelValue.value = icon
  close()
}

const handlePageChange = (page: number) => {
  setCurrentPage(page)
}

function toggleOpenState() {
  visible.value = !visible.value
}

// 监听 visible 变化，当打开时自动加载当前 Tab 的图标
watch(
  () => visible.value,
  (isOpen) => {
    if (isOpen && !sourceIconsMap.value[activeTab.value]?.length) {
      handleTabChange(activeTab.value)
    }
  }
)

function open() {
  visible.value = true
  // 打开时加载当前激活 Tab 的图标
  handleTabChange(activeTab.value)
}

function close() {
  visible.value = false
}

function onKeywordChange(v: string) {
  keyword.value = v
}

const searchInputProps = computed(() => {
  return {
    placeholder: $t('ui.iconPicker.search'),
    [props.modelValueProp]: keyword.value,
    [`onUpdate:${props.modelValueProp}`]: onKeywordChange,
    class: 'mx-2',
  }
})

function updateCurrentSelect(v: string) {
  currentSelect.value = v
  const eventKey = `onUpdate:${props.modelValueProp}`
  if (attrs[eventKey] && typeof attrs[eventKey] === 'function') {
    ;(attrs[eventKey] as Function)(v)
  }
}
const getBindAttrs = computed(() => {
  return objectOmit(attrs, [`onUpdate:${props.modelValueProp}`])
})

defineExpose({ toggleOpenState, open, close })
</script>

<template>
  <Popover
    v-model:open="visible"
    trigger="click"
    placement="bottomRight"
    overlay-class-name="enhanced-icon-picker-popover"
  >
    <template #content>
      <div class="w-full">
        <!-- Tab 切换栏 -->
        <Tabs
          v-if="availableSources.length > 1"
          :active-key="activeTab"
          size="small"
          class="enhanced-icon-picker-tabs"
          @change="(key: string) => handleTabChange(key)"
        >
          <TabPane
            v-for="source in availableSources"
            :key="source.key"
            :tab="source.label"
          />
        </Tabs>

        <!-- 搜索框 -->
        <div class="mb-2 flex w-full">
          <component
            v-if="inputComponent"
            :is="inputComponent"
            v-bind="searchInputProps"
          />
          <Input
            v-else
            class="mx-2 h-8 w-full"
            :placeholder="$t('ui.iconPicker.search')"
            v-model:value="keyword"
          />
        </div>

        <!-- 加载状态 -->
        <div
          v-if="sourceLoadingMap[activeTab]"
          class="flex flex-col items-center justify-center text-muted-foreground min-h-[150px] w-full"
        >
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <div class="mt-2 text-sm">加载中...</div>
        </div>

        <!-- 图标网格 -->
        <template v-else>
          <template v-if="paginationList.length > 0">
            <div class="grid max-h-[360px] w-full grid-cols-6 justify-items-center gap-1">
              <Tooltip
                v-for="(item, index) in paginationList"
                :key="index"
                :title="item"
                placement="top"
              >
                <Button
                  type="text"
                  size="small"
                  class="size-8 p-0!"
                  @click="handleClick(item)"
                >
                  <IconifyIcon
                    :icon="item"
                    :class="{
                      'text-primary': currentSelect === item,
                    }"
                    class="size-5"
                  />
                </Button>
              </Tooltip>
            </div>
            <div
              v-if="total >= pageSize"
              class="flex justify-end overflow-hidden border-t py-2 pr-3"
            >
              <Pagination
                :current="currentPage"
                :page-size="36"
                :simple="true"
                :total="total"
                size="small"
                @change="handlePageChange"
              />
            </div>
          </template>

          <!-- 空状态 -->
          <template v-else>
            <div class="flex flex-col items-center justify-center text-muted-foreground min-h-[150px] w-full">
              <EmptyIcon class="size-10" />
              <div class="mt-1 text-sm">{{ $t('common.noData') }}</div>
            </div>
          </template>
        </template>
      </div>
    </template>

    <!-- 触发器 -->
    <template v-if="props.type === 'input'">
      <component
        v-if="props.inputComponent"
        :is="inputComponent"
        :[modelValueProp]="currentSelect"
        :placeholder="$t('ui.iconPicker.placeholder')"
        role="combobox"
        :aria-label="$t('ui.iconPicker.placeholder')"
        :aria-expanded="visible"
        :[`onUpdate:${modelValueProp}`]="updateCurrentSelect"
        v-bind="getBindAttrs"
      >
        <template #[iconSlot]>
          <IconifyIcon
            :icon="currentSelect || Grip"
            class="size-4"
            aria-hidden="true"
          />
        </template>
      </component>
      <div class="relative w-full" v-else>
        <Input
          v-bind="$attrs"
          v-model:value="currentSelect"
          :placeholder="$t('ui.iconPicker.placeholder')"
          class="h-8 w-full pr-8"
          readonly
          role="combobox"
          :aria-label="$t('ui.iconPicker.placeholder')"
          :aria-expanded="visible"
          @click="visible = true"
        />
        <div
          class="absolute right-1 top-0 flex h-full items-center cursor-pointer"
          @click="visible = true"
        >
          <IconifyIcon
            v-if="currentSelect"
            :icon="currentSelect"
            class="size-5 text-primary"
            aria-hidden="true"
          />
          <IconifyIcon
            v-else
            :icon="Grip"
            class="size-5 text-muted-foreground hover:text-foreground transition-colors"
            aria-hidden="true"
          />
        </div>
      </div>
    </template>
    <IconifyIcon
      v-else
      :icon="currentSelect || Grip"
      class="size-4 cursor-pointer"
      v-bind="$attrs"
      @click="visible = true"
    />
  </Popover>
</template>

<style scoped>
.enhanced-icon-picker-popover {
  min-width: 320px;
  padding: 0;
  padding-top: 12px;
}

.enhanced-icon-picker-tabs {
  margin-bottom: 8px;
}
</style>
