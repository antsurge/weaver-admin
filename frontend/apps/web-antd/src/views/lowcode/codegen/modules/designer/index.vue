<script lang="ts" setup>
import type { FieldItem } from './field-lib';

import type { CodegenApi } from '#/api/lowcode/codegen';
import type { PermissionMenuApi } from '#/api/permission/menu';

import { computed, onMounted, ref } from 'vue';

import { IconifyIcon } from '@vben/icons';

import {
  Button,
  Card,
  Checkbox,
  Collapse,
  Empty,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  Select,
  Switch,
  Tag,
  TreeSelect,
} from 'ant-design-vue';

import {
  applyMenuApi,
  generateCodeApi,
  getDbColumnListApi,
} from '#/api/lowcode/codegen';
import { getMenuTreeApi } from '#/api/permission/menu';
import { $t } from '#/locales';

import { DRAG_MIME, FIELD_LIB } from './field-lib';

/**
 * 拖拽式 CRUD 设计器
 * mode=create：新建页面（支持传入默认数据：{ table } 选定的数据表，自动加载字段）
 * mode=edit  ：编辑已有 CRUD 记录（加载 fieldsJson）
 */
const props = defineProps<{
  /**
   * 默认数据（prop 传入）：
   * - { table: DbTable }：从数据表新建，预填表信息并自动加载字段到画布
   * - GenTable（含 id）：编辑已有 CRUD 记录，加载 fieldsJson
   */
  initialData?: CodegenApi.GenTable | null | { table?: CodegenApi.DbTable };
  mode: 'create' | 'edit';
}>();

const emit = defineEmits<{
  cancel: [];
  success: [id: string];
}>();

// ===== 基础信息 =====
const baseForm = ref({
  id: '',
  tableName: '',
  tableComment: '',
  moduleName: 'system',
  bizName: '',
  genType: 'table',
  menuEnabled: false,
  menuModule: '',
  buttons: [] as string[],
});
const advancedOpen = ref(false);

/** 菜单所属模块 TreeSelect 选中的节点 id（value 为唯一 id，模块名回显/提交时映射） */
const menuModuleId = ref('');

/** 可生成的按钮选项（与后端 GenBtn* 常量一一对应） */
const BTN_OPTIONS = [
  { label: '列表', value: 'list' },
  { label: '详情', value: 'detail' },
  { label: '创建', value: 'create' },
  { label: '编辑', value: 'edit' },
  { label: '删除', value: 'delete' },
  { label: '批量删除', value: 'batchDelete' },
  { label: '状态', value: 'status' },
  { label: '导入', value: 'import' },
  { label: '导出', value: 'export' },
];
/** 默认按钮全选 */
const ALL_BUTTONS = BTN_OPTIONS.map((b) => b.value);

// ===== 菜单树（上级目录选择） =====
interface MenuTreeNode {
  value: string;
  title: string;
  /** 归属模块名（节点 path 第一段，唯一性由 id 保证，模块名用于生成菜单落库） */
  module: string;
  disabled?: boolean;
  children?: MenuTreeNode[];
}
const menuTreeData = ref<MenuTreeNode[]>([]);
const loadingMenuTree = ref(false);

/** 把目录节点的小写路径作为模块名（如 path=/security -> security） */
function toModuleName(node: PermissionMenuApi.PermissionMenu): string {
  if (node.path) return node.path.replace(/^\/+/, '').split('/')[0] || '';
  if (node.authCode) return node.authCode.toLowerCase();
  return node.code || node.name || '';
}

/** 在树中按模块名查找节点 id（用于编辑回显：模块名 -> 节点 id） */
function resolveModuleId(tree: MenuTreeNode[], module: string): string {
  for (const node of tree) {
    if (node.module === module) return node.value;
    if (node.children?.length) {
      const id = resolveModuleId(node.children, module);
      if (id) return id;
    }
  }
  return '';
}

/** 在树中按节点 id 查找模块名（提交时映射：节点 id -> 模块名） */
function resolveModuleById(tree: MenuTreeNode[], id: string): string {
  for (const node of tree) {
    if (node.value === id) return node.module;
    if (node.children?.length) {
      const module = resolveModuleById(node.children, id);
      if (module) return module;
    }
  }
  return '';
}

/** 菜单模块 TreeSelect 清空（allow-clear 点击清除时同步清空模块名） */
function onMenuModuleChange(_: string) {
  if (!_) baseForm.value.menuModule = '';
}

/** 递归收集菜单树中可选的目录节点（仅 catalog 可作为上级菜单选择，menu/action 不可选） */
async function loadMenuTree() {
  loadingMenuTree.value = true;
  try {
    const res = await getMenuTreeApi();
    const items = (res?.items ?? []) as PermissionMenuApi.PermissionMenu[];
    menuTreeData.value = items
      .filter((n) => n.type === 'catalog')
      .map((n) => toTreeNode(n));
    // 编辑回显：菜单所属模块已有值（模块名）时反查节点 id
    if (baseForm.value.menuModule) {
      menuModuleId.value = resolveModuleId(
        menuTreeData.value,
        baseForm.value.menuModule,
      );
    }
  } catch {
    menuTreeData.value = [];
  } finally {
    loadingMenuTree.value = false;
  }
}

/** 递归构建 TreeSelect 树节点：仅 catalog 可选，menu 节点禁用仅作层级上下文展示。
 * value 必须用唯一 id（模块名会被同一目录下多个节点共享，导致 TreeSelect 渲染重复）。 */
function toTreeNode(node: PermissionMenuApi.PermissionMenu): MenuTreeNode {
  const isCatalog = node.type === 'catalog';
  const children = (node.children ?? [])
    .filter((c) => c.type === 'catalog' || c.type === 'menu')
    .map((c) => toTreeNode(c));
  return {
    value: node.id,
    module: toModuleName(node),
    // 展示菜单标题（title 为多语言 key，翻译后即中文名），与菜单管理"上级菜单"一致
    title: $t(node.title || node.name),
    disabled: !isCatalog,
    children: children.length > 0 ? children : undefined,
  };
}

// 当前选中字段 key
const activeKey = ref('');
const dragging = ref(false);

const fieldKeySeq = ref(0);
const fields = ref<FieldItem[]>([]);

const loadingFields = ref(false);
const emitting = ref(false);
const generating = ref(false);

const COMPONENT_TYPES = [
  'Input',
  'Textarea',
  'Select',
  'Radio',
  'Checkbox',
  'DatePicker',
  'InputNumber',
  'Switch',
];

const GO_TYPES = ['string', 'int64', 'float64', 'bool', 'time.Time'];

const activeField = computed(() =>
  fields.value.find((f) => f.key === activeKey.value),
);

/** 左右折叠：字段库 / 画布 / 属性 */
const libCollapsedGroups = ref<string[]>([]);

function toCamel(name: string): string {
  const camel = name.replaceAll(/_([a-z])/g, (_, c) => c.toUpperCase());
  return camel.charAt(0).toUpperCase() + camel.slice(1);
}

function mapGoType(dataType: string): string {
  switch (dataType) {
    case 'bigint':
    case 'int':
    case 'mediumint':
    case 'smallint':
    case 'tinyint': {
      return 'int64';
    }
    case 'date':
    case 'datetime':
    case 'time':
    case 'timestamp': {
      return 'time.Time';
    }
    case 'decimal':
    case 'double':
    case 'float': {
      return 'float64';
    }
    default: {
      return 'string';
    }
  }
}

function mapComponentType(dataType: string): string {
  switch (dataType) {
    case 'bigint':
    case 'decimal':
    case 'double':
    case 'float':
    case 'int':
    case 'mediumint':
    case 'smallint':
    case 'tinyint': {
      return 'InputNumber';
    }
    case 'date':
    case 'datetime':
    case 'time':
    case 'timestamp': {
      return 'DatePicker';
    }
    case 'longtext':
    case 'mediumtext':
    case 'text': {
      return 'Textarea';
    }
    default: {
      return 'Input';
    }
  }
}

// ===== 字段库 → 画布 =====

/** 由下拉列表查找到字段类型定义 */
function findLibField(type: string) {
  for (const g of FIELD_LIB) {
    const f = g.fields.find((item) => item.type === type);
    if (f) return f;
  }
  return undefined;
}

/**
 * 列名去重：首个字段直接用前缀（status），重复时追加 _01、_02…（两位数字），
 * 与系统字段命名保持一致（id、status、created_at 均不带序号）。
 */
function nextColumnName(prefix: string): string {
  const used = new Set(fields.value.map((f) => f.columnName));
  if (!used.has(prefix)) return prefix;
  let i = 1;
  while (used.has(`${prefix}_${String(i).padStart(2, '0')}`)) {
    i++;
  }
  return `${prefix}_${String(i).padStart(2, '0')}`;
}

/** 生成一个画布字段项（列名/注释自动去重：首个不加后缀，重复时追加 _01、_02…） */
function createFieldItem(type: string): FieldItem {
  const lib = findLibField(type);
  const seq = ++fieldKeySeq.value;
  const tpl = lib
    ? lib.template
    : {
        colPrefix: 'field',
        columnComment: '字段',
        goType: 'string' as const,
        componentType: 'Input' as const,
        length: 255,
        primaryKey: false,
        list: true,
        form: true,
        query: false,
        required: false,
      };
  const columnName = nextColumnName(tpl.colPrefix);
  // 列名带后缀（status_01）时注释同步带后缀（状态_01），否则保持模板注释（状态）
  const suffix = columnName.slice(tpl.colPrefix.length);
  const fieldName = toCamel(columnName);
  return {
    key: `f_${Date.now()}_${seq}`,
    columnName,
    columnComment: `${tpl.columnComment}${suffix}`,
    fieldName,
    goType: tpl.goType,
    componentType: tpl.componentType,
    length: tpl.length,
    primaryKey: tpl.primaryKey,
    list: tpl.list,
    form: tpl.form,
    query: tpl.query,
    required: tpl.required,
  };
}

/** 拖拽：左侧字段项开始拖 */
function onLibDragStart(e: DragEvent, type: string) {
  dragging.value = true;
  e.dataTransfer?.setData(DRAG_MIME, type);
  e.dataTransfer!.effectAllowed = 'copy';
}

/** 模拟点击添加（兼容拖拽不可用环境） */
function onClickAdd(type: string) {
  const item = createFieldItem(type);
  fields.value.push(item);
  activeKey.value = item.key;
}

/** 拖拽：画布接受 */
function onCanvasDragOver(e: DragEvent) {
  e.preventDefault();
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
}

function onCanvasDrop(e: DragEvent) {
  e.preventDefault();
  dragging.value = false;
  const type = e.dataTransfer?.getData(DRAG_MIME);
  if (!type) return;
  const item = createFieldItem(type);
  fields.value.push(item);
  activeKey.value = item.key;
}

function onDragEnd() {
  dragging.value = false;
}

// ===== 画布字段操作 =====
function removeField(key: string) {
  fields.value = fields.value.filter((f) => f.key !== key);
  if (activeKey.value === key) activeKey.value = '';
}

function moveField(key: string, dir: -1 | 1) {
  const idx = fields.value.findIndex((f) => f.key === key);
  const target = idx + dir;
  if (idx < 0 || target < 0 || target >= fields.value.length) return;
  const arr = [...fields.value];
  const a = arr[idx]!;
  const b = arr[target]!;
  arr[idx] = b;
  arr[target] = a;
  fields.value = arr;
}

/** 画布内拖拽排序（HTML5 DnD 同一 DataTransfer） */
function onFieldDragStart(e: DragEvent, key: string) {
  e.dataTransfer?.setData('text/plain', key);
  e.dataTransfer!.effectAllowed = 'move';
}

function onFieldDrop(e: DragEvent, targetKey: string) {
  e.preventDefault();
  const fromKey = e.dataTransfer?.getData('text/plain');
  if (!fromKey || fromKey === targetKey) return;
  const fromIdx = fields.value.findIndex((f) => f.key === fromKey);
  const toIdx = fields.value.findIndex((f) => f.key === targetKey);
  if (fromIdx === -1 || toIdx === -1) return;
  const arr = [...fields.value];
  const [moved] = arr.splice(fromIdx, 1);
  arr.splice(toIdx, 0, moved!);
  fields.value = arr;
}

// ===== 属性配置 =====

/** 列名变更时同步生成字段名 */
function onColumnNameChange(field: FieldItem) {
  field.fieldName = toCamel(field.columnName || '');
}

// ===== 初始化（按模式加载） =====
async function init() {
  fieldKeySeq.value = 0;
  fields.value = [];
  activeKey.value = '';
  loadingFields.value = false;
  generating.value = false;
  advancedOpen.value = false;

  if (props.mode === 'edit') {
    const data = props.initialData as CodegenApi.GenTable;
    if (!data) return;
    applyBaseForm(data);
    if (data.fieldsJson?.fields) {
      fields.value = (data.fieldsJson.fields as any[]).map(
        (f: any, idx: number) => ({
          key: `f_${Date.now()}_${idx}`,
          columnName: f.columnName || '',
          columnComment: f.columnComment || '',
          fieldName: f.fieldName || '',
          goType: f.goType || 'string',
          componentType: f.componentType || 'Input',
          length: f.length || 0,
          primaryKey: !!f.primaryKey,
          list: f.list ?? true,
          form: f.form ?? true,
          query: !!f.query,
          required: !!f.required,
        }),
      );
      fieldKeySeq.value = fields.value.length;
      activeKey.value = fields.value[0]?.key || '';
    }
    return;
  }

  // create：新建页面（可选：传入 { table } 默认数据预填并加载字段）
  baseForm.value = {
    id: '',
    tableName: '',
    tableComment: '',
    moduleName: 'system',
    bizName: '',
    genType: 'table',
    menuEnabled: false,
    menuModule: '',
    buttons: [...ALL_BUTTONS],
  };
  const data = props.initialData as null | { table?: CodegenApi.DbTable };
  if (data?.table) {
    baseForm.value.tableName = data.table.tableName;
    baseForm.value.tableComment = data.table.tableComment || '';
    baseForm.value.bizName = toCamel(data.table.tableName);
    await loadDbColumns();
  }
}

function applyBaseForm(data: CodegenApi.GenTable) {
  baseForm.value = {
    id: data.id,
    tableName: data.tableName,
    tableComment: data.tableComment || '',
    moduleName: data.moduleName || 'system',
    bizName: data.bizName || '',
    genType: data.genType || 'table',
    menuEnabled: data.menuEnabled ?? false,
    menuModule: data.menuModule || '',
    buttons: data.buttons?.length ? data.buttons : [...ALL_BUTTONS],
  };
}

async function loadDbColumns() {
  if (!baseForm.value.tableName) return;
  loadingFields.value = true;
  try {
    const res = await getDbColumnListApi(baseForm.value.tableName);
    const cols: CodegenApi.DbColumn[] = res?.items ?? [];
    fields.value = cols.map((c: CodegenApi.DbColumn, idx: number) => ({
      key: `f_${Date.now()}_${idx}`,
      columnName: c.columnName,
      columnComment: c.columnComment || c.columnName,
      fieldName: toCamel(c.columnName),
      goType: mapGoType(c.dataType),
      componentType: mapComponentType(c.dataType),
      length: 0,
      primaryKey: c.isPrimaryKey,
      list: !c.isPrimaryKey,
      form: !c.isPrimaryKey,
      query: false,
      required: !c.isNullable,
    }));
    fieldKeySeq.value = fields.value.length;
    if (cols.length === 0) {
      message.warning($t('lowcode.codegen.message.noColumn'));
    }
  } catch (error) {
    message.error(
      (error as Error)?.message || $t('lowcode.codegen.message.loadColumnFail'),
    );
  } finally {
    loadingFields.value = false;
  }
}

onMounted(() => {
  init();
  loadMenuTree();
});

// ===== 保存 & 生成 =====

const canGenerate = computed(() => {
  const t = baseForm.value.tableName.trim();
  const m = baseForm.value.moduleName.trim();
  const b = baseForm.value.bizName.trim();
  return (
    t !== '' &&
    m !== '' &&
    b !== '' &&
    fields.value.length > 0 &&
    fields.value.every((f) => f.columnName?.trim())
  );
});

/** 生成 CRUD 代码：一次调用（保存配置 + 生成代码），随后展示预览 */
async function onGenerate() {
  if (!canGenerate.value) {
    message.warning($t('lowcode.codegen.message.incomplete'));
    return;
  }
  emitting.value = true;
  try {
    // 菜单所属模块：优先以选中节点 id 反查模块名（value 为唯一 id）
    const resolvedModule = menuModuleId.value
      ? resolveModuleById(menuTreeData.value, menuModuleId.value)
      : '';
    const payload = {
      id: baseForm.value.id || undefined,
      tableName: baseForm.value.tableName,
      tableComment: baseForm.value.tableComment,
      moduleName: baseForm.value.moduleName,
      bizName: baseForm.value.bizName,
      genType: baseForm.value.genType,
      status: 'enabled',
      genMenu: baseForm.value.menuEnabled,
      menuModule: resolvedModule || baseForm.value.menuModule,
      buttons: baseForm.value.buttons,
      fieldsJson: { fields: fields.value },
    } as any;

    generating.value = true;
    const res = await generateCodeApi(payload);
    const id = res?.id || '';
    if (id) {
      baseForm.value.id = id;
      emit('success', id);
    }
    previewFiles.value = res?.files || [];
    activeFile.value = previewFiles.value[0]?.fileName || '';
    if (previewFiles.value.length === 0) {
      message.warning($t('lowcode.codegen.message.generateEmpty'));
    }
    // 开启生成菜单时，生成成功后自动落库菜单（接口幂等，重复点击安全）
    if (baseForm.value.menuEnabled && id) {
      try {
        await applyMenuApi(id);
      } catch (error) {
        // 代码已生成，菜单落库失败不阻断，仅提示
        message.error(
          (error as Error)?.message ||
            $t('lowcode.codegen.message.menuApplyFail'),
        );
      }
    }
  } finally {
    emitting.value = false;
    generating.value = false;
  }
}

function onCancel() {
  emit('cancel');
}

const applyingMenu = ref(false);

/** 一键落库生成菜单（调用后端菜单创建方法） */
async function onApplyMenu() {
  if (!baseForm.value.id) {
    message.warning($t('lowcode.codegen.message.saveFirst'));
    return;
  }
  applyingMenu.value = true;
  try {
    await applyMenuApi(baseForm.value.id);
  } finally {
    applyingMenu.value = false;
  }
}

// ===== 代码预览 =====
const previewFiles = ref<CodegenApi.GenCodeFile[]>([]);
const activeFile = ref('');
const showPreview = computed(() => previewFiles.value.length > 0);
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <!-- ============ 顶部操作栏 ============ -->
    <div class="border-b border-gray-200 pb-3 dark:border-gray-700">
      <div class="flex flex-wrap items-center gap-3">
        <div class="text-lg font-semibold">
          {{ $t('lowcode.codegen.designer.title') }}
        </div>
        <div class="flex items-center gap-1">
          <span class="text-red-500">*</span>
          <Input
            v-model:value="baseForm.tableName"
            class="w-48"
            :placeholder="$t('lowcode.codegen.designer.tableNamePlaceholder')"
          />
        </div>
        <Input
          v-model:value="baseForm.tableComment"
          class="w-60"
          :placeholder="$t('lowcode.codegen.designer.tableCommentPlaceholder')"
        />
        <div class="flex items-center gap-1">
          <Button
            size="small"
            type="link"
            @click="advancedOpen = !advancedOpen"
          >
            <IconifyIcon class="mr-1" icon="ant-design:setting-outlined" />
            {{ $t('lowcode.codegen.designer.advanced') }}
          </Button>
          <IconifyIcon
            class="cursor-pointer text-gray-400"
            :icon="
              advancedOpen
                ? 'ant-design:up-outlined'
                : 'ant-design:down-outlined'
            "
            @click="advancedOpen = !advancedOpen"
          />
        </div>
        <div class="ml-auto flex items-center gap-2">
          <Button
            type="primary"
            size="middle"
            :loading="emitting || generating"
            :disabled="!canGenerate"
            @click="onGenerate"
          >
            <IconifyIcon class="mr-1" icon="ant-design:rocket-outlined" />
            {{ $t('lowcode.codegen.designer.generateCrud') }}
          </Button>
          <Button danger size="middle" @click="onCancel">
            {{ $t('lowcode.codegen.designer.giveUp') }}
          </Button>
        </div>
      </div>
      <!-- 高级配置默认隐藏，故在此给出提示描述（需求3） -->
      <div
        v-if="!advancedOpen"
        class="mt-2 flex items-center gap-1 text-xs text-gray-400 dark:text-gray-500"
      >
        <IconifyIcon icon="ant-design:info-circle-outlined" />
        <span>{{ $t('lowcode.codegen.designer.advancedTip') }}</span>
        <a
          class="ml-1 cursor-pointer text-blue-500 hover:underline"
          @click="advancedOpen = true"
        >
          {{ $t('lowcode.codegen.designer.expandAdvanced') }}
        </a>
      </div>
    </div>

    <!-- 高级配置 -->
    <div
      v-if="advancedOpen"
      class="mt-2 grid grid-cols-3 gap-3 rounded border border-dashed border-gray-300 p-3 dark:border-gray-600"
    >
      <div>
        <label class="mb-1 block text-sm text-gray-600 dark:text-gray-300">
          <span class="text-red-500">* </span>
          {{ $t('lowcode.codegen.fields.moduleName') }}
        </label>
        <Input
          v-model:value="baseForm.moduleName"
          class="w-full"
          :placeholder="$t('lowcode.codegen.designer.moduleNamePlaceholder')"
        />
        <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
          {{ $t('lowcode.codegen.designer.moduleNameDesc') }}
        </p>
      </div>
      <div>
        <label class="mb-1 block text-sm text-gray-600 dark:text-gray-300">
          <span class="text-red-500">* </span>
          {{ $t('lowcode.codegen.fields.bizName') }}
        </label>
        <Input
          v-model:value="baseForm.bizName"
          class="w-full"
          :placeholder="$t('lowcode.codegen.designer.bizNamePlaceholder')"
        />
      </div>
      <div>
        <label class="mb-1 block text-sm text-gray-600 dark:text-gray-300">
          {{ $t('lowcode.codegen.fields.genType') }}
        </label>
        <Select
          v-model:value="baseForm.genType"
          class="w-full"
          :placeholder="$t('lowcode.codegen.designer.genTypePlaceholder')"
          :options="[
            { label: $t('lowcode.codegen.gen_types.table'), value: 'table' },
            { label: $t('lowcode.codegen.gen_types.form'), value: 'form' },
          ]"
        />
      </div>
      <div>
        <label class="mb-1 block text-sm text-gray-600 dark:text-gray-300">
          {{ $t('lowcode.codegen.fields.menuEnabled') }}
        </label>
        <Switch v-model:checked="baseForm.menuEnabled" />
      </div>
      <div>
        <label class="mb-1 block text-sm text-gray-600 dark:text-gray-300">
          {{ $t('lowcode.codegen.fields.menuModule') }}
        </label>
        <TreeSelect
          v-model:value="menuModuleId"
          class="w-full"
          :tree-data="menuTreeData"
          :placeholder="$t('lowcode.codegen.designer.menuModulePlaceholder')"
          :loading="loadingMenuTree"
          allow-clear
          show-search
          tree-default-expand-all
          tree-node-filter-prop="title"
          :max-tag-count="4"
          @change="onMenuModuleChange"
        />
        <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
          {{ $t('lowcode.codegen.designer.menuModuleDesc') }}
        </p>
      </div>
      <div>
        <label class="mb-1 block text-sm text-gray-600 dark:text-gray-300">
          {{ $t('lowcode.codegen.fields.buttons') }}
        </label>
        <Select
          v-model:value="baseForm.buttons"
          class="w-full"
          mode="multiple"
          :options="BTN_OPTIONS"
          :max-tag-count="4"
          :placeholder="$t('lowcode.codegen.designer.buttonsPlaceholder')"
        />
      </div>
      <div
        class="col-span-3 flex items-center gap-2 border-t border-dashed border-gray-200 pt-2 dark:border-gray-700"
      >
        <span class="text-sm text-gray-500">
          {{ $t('lowcode.codegen.designer.menuTip') }}
        </span>
        <Button
          size="small"
          type="primary"
          ghost
          :loading="applyingMenu"
          :disabled="!baseForm.id || !baseForm.menuEnabled"
          @click="onApplyMenu"
        >
          <IconifyIcon class="mr-1" icon="ant-design:apartment-outlined" />
          {{ $t('lowcode.codegen.operation.applyMenu') }}
        </Button>
      </div>
    </div>

    <!-- ============ 三栏主体 ============ -->
    <div class="mt-3 flex min-h-0 flex-1 gap-3">
      <!-- 左侧：字段库 -->
      <div
        class="w-56 shrink-0 overflow-auto rounded border border-gray-200 bg-white p-2 dark:border-gray-700 dark:bg-gray-800"
      >
        <Collapse
          v-model:active-key="libCollapsedGroups"
          ghost
          class="codegen-lib-collapse"
        >
          <Collapse.Panel
            v-for="g in FIELD_LIB"
            :key="g.group"
            :header="$t(g.i18nKey)"
          >
            <div class="flex flex-col gap-1">
              <div
                v-for="f in g.fields"
                :key="f.type"
                class="cursor-pointer rounded border border-gray-200 px-2 py-1 text-sm hover:border-blue-400 hover:text-blue-500 dark:border-gray-600"
                draggable="true"
                @dragstart="onLibDragStart($event, f.type)"
                @dragend="onDragEnd"
                @click="onClickAdd(f.type)"
              >
                {{ f.label }}
              </div>
            </div>
            <div class="mt-1 text-center text-xs text-gray-300">
              {{ $t('lowcode.codegen.designer.dragHint') }}
            </div>
          </Collapse.Panel>
        </Collapse>
      </div>

      <!-- 中间：画布 -->
      <div
        class="flex-1 overflow-auto rounded border-2 border-dashed p-3 transition"
        :class="
          dragging
            ? 'border-blue-400 bg-blue-50/50 dark:bg-blue-900/10'
            : 'border-gray-200 dark:border-gray-700'
        "
        @dragover="onCanvasDragOver"
        @drop="onCanvasDrop"
        @dragend="onDragEnd"
      >
        <div
          v-if="fields.length === 0"
          class="flex h-full min-h-[260px] items-center justify-center"
        >
          <Empty :description="$t('lowcode.codegen.designer.canvasEmpty')" />
        </div>
        <template v-else>
          <Card
            v-for="field in fields"
            :key="field.key"
            size="small"
            class="codegen-canvas-card mb-2"
            :class="{ 'ring-2 ring-blue-400': activeKey === field.key }"
            :draggable="true"
            @click="activeKey = field.key"
            @dragstart="onFieldDragStart($event, field.key)"
            @dragover.prevent
            @drop="onFieldDrop($event, field.key)"
          >
            <template #title>
              <div class="flex items-center gap-2">
                <IconifyIcon
                  class="text-gray-400"
                  icon="ant-design:holder-outlined"
                />
                <span>{{ field.columnComment }}</span>
                <Tag v-if="field.primaryKey" color="red">PK</Tag>
                <Tag color="blue">{{ field.componentType }}</Tag>
              </div>
            </template>
            <template #extra>
              <div class="flex gap-1">
                <Button
                  size="small"
                  type="text"
                  @click.stop="moveField(field.key, -1)"
                >
                  <IconifyIcon icon="ant-design:up-outlined" />
                </Button>
                <Button
                  size="small"
                  type="text"
                  @click.stop="moveField(field.key, 1)"
                >
                  <IconifyIcon icon="ant-design:down-outlined" />
                </Button>
                <Button
                  size="small"
                  type="text"
                  danger
                  @click.stop="removeField(field.key)"
                >
                  <IconifyIcon icon="ant-design:delete-outlined" />
                </Button>
              </div>
            </template>
            <div
              class="flex flex-wrap items-center gap-3 text-sm text-gray-500 dark:text-gray-400"
            >
              <span class="font-mono text-gray-700 dark:text-gray-200">{{
                field.columnName
              }}</span>
              <span v-if="(field.length ?? 0) > 0"
                >长度 {{ field.length }}</span
              >
              <span
                >{{ $t('lowcode.codegen.field_config.list') }}:
                <Checkbox
                  :checked="field.list"
                  :disabled="field.primaryKey"
                  @change="(e: any) => (field.list = e.target.checked)"
                />
              </span>
              <span
                >{{ $t('lowcode.codegen.field_config.form') }}:
                <Checkbox
                  :checked="field.form"
                  :disabled="field.primaryKey"
                  @change="(e: any) => (field.form = e.target.checked)"
                />
              </span>
              <span
                >{{ $t('lowcode.codegen.field_config.query') }}:
                <Checkbox
                  :checked="field.query"
                  :disabled="field.primaryKey"
                  @change="(e: any) => (field.query = e.target.checked)"
                />
              </span>
              <span
                >{{ $t('lowcode.codegen.field_config.required') }}:
                <Checkbox
                  :checked="field.required"
                  :disabled="field.primaryKey"
                  @change="(e: any) => (field.required = e.target.checked)"
                />
              </span>
            </div>
          </Card>
        </template>
      </div>

      <!-- 右侧：属性配置 -->
      <div
        class="w-72 shrink-0 overflow-auto rounded border border-gray-200 bg-white p-3 dark:border-gray-700 dark:bg-gray-800"
      >
        <div class="mb-3 font-medium">
          {{ $t('lowcode.codegen.designer.properties') }}
        </div>
        <template v-if="activeField">
          <Form layout="vertical" size="small">
            <FormItem :label="$t('lowcode.codegen.fields.columnName')">
              <Input
                v-model:value="activeField.columnName"
                :placeholder="
                  $t('lowcode.codegen.designer.columnNamePlaceholder')
                "
                @change="onColumnNameChange(activeField)"
              />
            </FormItem>
            <FormItem :label="$t('lowcode.codegen.fields.columnComment')">
              <Input
                v-model:value="activeField.columnComment"
                :placeholder="
                  $t('lowcode.codegen.designer.columnCommentPlaceholder')
                "
              />
            </FormItem>
            <FormItem :label="$t('lowcode.codegen.fields.fieldName')">
              <Input
                v-model:value="activeField.fieldName"
                :placeholder="
                  $t('lowcode.codegen.designer.fieldNamePlaceholder')
                "
              />
            </FormItem>
            <FormItem :label="$t('lowcode.codegen.field_config.componentType')">
              <Select
                v-model:value="activeField.componentType"
                :options="COMPONENT_TYPES.map((t) => ({ label: t, value: t }))"
              />
            </FormItem>
            <FormItem :label="$t('lowcode.codegen.field_config.goType')">
              <Select
                v-model:value="activeField.goType"
                :options="GO_TYPES.map((t) => ({ label: t, value: t }))"
              />
            </FormItem>
            <FormItem :label="$t('lowcode.codegen.designer.length')">
              <InputNumber
                v-model:value="activeField.length"
                :min="0"
                class="w-full"
                :placeholder="$t('lowcode.codegen.designer.lengthPlaceholder')"
              />
            </FormItem>
            <FormItem :label="$t('lowcode.codegen.designer.flags')">
              <div class="flex flex-col gap-1">
                <Checkbox v-model:checked="activeField.primaryKey">
                  {{ $t('lowcode.codegen.designer.isPrimaryKey') }}
                </Checkbox>
                <Checkbox v-model:checked="activeField.required">
                  {{ $t('lowcode.codegen.field_config.required') }}
                </Checkbox>
                <Checkbox v-model:checked="activeField.list">
                  {{ $t('lowcode.codegen.field_config.list') }}
                </Checkbox>
                <Checkbox v-model:checked="activeField.form">
                  {{ $t('lowcode.codegen.field_config.form') }}
                </Checkbox>
                <Checkbox v-model:checked="activeField.query">
                  {{ $t('lowcode.codegen.field_config.query') }}
                </Checkbox>
              </div>
            </FormItem>
          </Form>
        </template>
        <div v-else class="text-sm text-gray-400">
          {{ $t('lowcode.codegen.designer.noConfig') }}
        </div>
      </div>
    </div>

    <!-- ============ 生成结果预览（抽屉式） ============ -->
    <div
      v-if="showPreview"
      class="mt-3 flex h-80 flex-col rounded border border-gray-200 dark:border-gray-700"
    >
      <div
        class="flex items-center gap-2 border-b border-gray-200 px-3 py-2 dark:border-gray-700"
      >
        <span class="font-medium">{{
          $t('lowcode.codegen.designer.generatedResult')
        }}</span>
        <span class="text-sm text-gray-500"
          >{{ previewFiles.length }}
          {{ $t('lowcode.codegen.designer.files') }}</span
        >
        <div class="ml-auto flex gap-2">
          <Button size="small" :loading="generating" @click="onGenerate">
            {{ $t('lowcode.codegen.operation.generate') }}
          </Button>
          <Button size="small" type="primary" ghost @click="onCancel">
            <IconifyIcon class="mr-1" icon="ant-design:home-outlined" />
            {{ $t('lowcode.codegen.designer.backToStart') }}
          </Button>
        </div>
      </div>
      <div class="flex min-h-0 flex-1 gap-2 p-2">
        <div
          class="w-60 flex-shrink-0 overflow-auto border-r border-gray-200 pr-2 dark:border-gray-700"
        >
          <div
            v-for="file in previewFiles"
            :key="file.fileName"
            class="mb-1 cursor-pointer truncate rounded px-2 py-1 text-xs"
            :class="
              activeFile === file.fileName
                ? 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-400'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'
            "
            @click="activeFile = file.fileName"
          >
            {{ file.fileName }}
          </div>
        </div>
        <pre
          class="flex-1 overflow-auto rounded bg-gray-800 p-3 text-xs text-gray-100"
          >{{
            previewFiles.find((f) => f.fileName === activeFile)?.content
          }}</pre>
      </div>
    </div>
  </div>
</template>
