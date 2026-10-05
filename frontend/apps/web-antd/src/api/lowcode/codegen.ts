import type {
  AllResult,
  PaginationParams,
  PaginationResult,
} from '#/types/pagination';

import { requestClient } from '#/api/request';

export namespace CodegenApi {
  /** 代码生成配置 */
  export interface GenTable {
    /** 配置ID */
    id: string;
    /** 数据库表名 */
    tableName: string;
    /** 表备注 */
    tableComment?: string;
    /** 模块名（生成代码的目录名，如 system） */
    moduleName?: string;
    /** 业务名（驼峰，如 systemUser） */
    bizName?: string;
    /** 字段配置（JSON，结构见 GenField） */
    fieldsJson?: Record<string, unknown>;
    /** 生成类型：table=整表生成 form=表单生成 */
    genType?: string;
    /** 是否生成菜单 */
    menuEnabled?: boolean;
    /** 菜单所属模块（父目录归属） */
    menuModule?: string;
    /** 生成的按钮列表 */
    buttons?: string[];
    /** 状态 */
    status?: string;
    /** 创建时间 */
    createdAt?: string;
    /** 更新时间 */
    updatedAt?: string;
  }

  /** 生成字段配置 */
  export interface GenField {
    /** 数据库列名 */
    columnName: string;
    /** 列注释 */
    columnComment: string;
    /** 字段名（驼峰） */
    fieldName?: string;
    /** Go 类型 */
    goType?: string;
    /** 前端组件类型：Input/Select/DatePicker/... */
    componentType?: string;
    /** 是否列表展示 */
    list?: boolean;
    /** 是否表单展示 */
    form?: boolean;
    /** 是否必填 */
    required?: boolean;
    /** 是否主键 */
    primaryKey?: boolean;
    /** 是否查询条件 */
    query?: boolean;
    /** 字段长度（字符串/整型宽度，0 表示不限制） */
    length?: number;
  }

  export interface GenTableListParams extends PaginationParams {
    tableName?: string;
    bizName?: string;
    status?: string;
  }

  /** 数据库表元数据 */
  export interface DbTable {
    tableName: string;
    tableComment?: string;
  }

  /** 数据库列元数据 */
  export interface DbColumn {
    columnName: string;
    columnComment?: string;
    dataType: string;
    isPrimaryKey: boolean;
    isNullable: boolean;
    autoIncrement: boolean;
  }

  /** 生成的代码文件 */
  export interface GenCodeFile {
    fileName: string;
    content: string;
  }
}

/**
 * 获取代码生成配置列表
 */
async function getGenTableListApi(params?: CodegenApi.GenTableListParams) {
  return requestClient.get<PaginationResult<CodegenApi.GenTable>>(
    '/admin/v1/gen-tables',
    { params },
  );
}

/**
 * 获取数据库表列表（后端返回 { items: DbTable[] }）
 */
async function getDbTableListApi(keyword?: string) {
  return requestClient.get<AllResult<CodegenApi.DbTable>>(
    '/admin/v1/codegen/db-tables',
    { params: keyword ? { keyword } : undefined },
  );
}

/**
 * 获取指定表的字段列表
 * 注意：后端返回结构为 { items: DbColumn[] }，与 ListDbTables 一致
 */
async function getDbColumnListApi(tableName: string) {
  return requestClient.get<{ items: CodegenApi.DbColumn[] }>(
    `/admin/v1/codegen/db-columns/${tableName}`,
  );
}

/**
 * 创建代码生成配置
 */
async function createGenTableApi(
  data: Omit<CodegenApi.GenTable, 'createdAt' | 'id' | 'updatedAt'>,
) {
  return requestClient.post('/admin/v1/gen-tables', data, {
    showSuccessMessage: true,
  });
}

/**
 * 删除代码生成配置记录（批量，仅删除配置记录，不影响已生成的代码文件）
 */
async function deleteGenTableApi(ids: string[]) {
  return requestClient.delete('/admin/v1/gen-tables', {
    params: { ids },
  });
}

/**
 * 彻底删除生成的模块（批量）：删除写盘的代码文件、删除真实数据表、
 * 删除已落库的菜单树、物理删除配置记录
 */
async function deleteGeneratedApi(ids: string[]) {
  return requestClient.post<{
    droppedTables?: string[];
    message?: string;
    removedFiles?: number;
  }>('/admin/v1/codegen/delete-generated', { ids });
}

/**
 * 生成代码（一次调用：保存配置 + 生成代码）
 * @param data 完整生成参数（含配置信息 + 字段配置），id 为空时后端自动新建配置
 * @param data.bizName 业务名（驼峰）
 * @param data.buttons 生成的按钮列表
 * @param data.fieldsJson 字段配置
 * @param data.genMenu 是否生成菜单
 * @param data.genType 生成类型
 * @param data.id 生成表配置 ID（为空则新建配置并落库）
 * @param data.menuModule 菜单所属模块
 * @param data.moduleName 模块名
 * @param data.status 状态
 * @param data.tableComment 表备注
 * @param data.tableName 数据库表名
 */
async function generateCodeApi(data: {
  /** 业务名（驼峰） */
  bizName?: string;
  /** 生成的按钮列表 */
  buttons?: string[];
  /** 字段配置 */
  fieldsJson?: Record<string, unknown>;
  /** 是否生成菜单 */
  genMenu?: boolean;
  /** 生成类型 */
  genType?: string;
  /** 生成表配置 ID（为空则新建配置并落库） */
  id?: string;
  /** 菜单所属模块 */
  menuModule?: string;
  /** 模块名 */
  moduleName?: string;
  /** 状态 */
  status?: string;
  /** 表备注 */
  tableComment?: string;
  /** 数据库表名 */
  tableName: string;
}) {
  return requestClient.post<{ files: CodegenApi.GenCodeFile[]; id: string }>(
    '/admin/v1/codegen/generate',
    data,
    { showSuccessMessage: true },
  );
}

/**
 * 一键落库生成菜单树（调用后端菜单创建方法）
 */
async function applyMenuApi(id: string) {
  return requestClient.post(
    '/admin/v1/codegen/apply-menu',
    { id },
    {
      showSuccessMessage: true,
    },
  );
}

export {
  applyMenuApi,
  createGenTableApi,
  deleteGeneratedApi,
  deleteGenTableApi,
  generateCodeApi,
  getDbColumnListApi,
  getDbTableListApi,
  getGenTableListApi,
};
