/**
 * 单个语言项类型
 */
export interface LocaleOption {
  /** 语言标识，对应 i18n key，例如 'zh-CN', 'en-US' */
  value: string;
  /** 显示名称，例如 '中文简体', 'English' */
  label: string;
}

/**
 * 类型推断
 * LocaleItem = { value: "zh-CN"; label: "中文简体" } | { value: "en-US"; label: "English" }
 */
export type LocaleItem = LocaleOption;

/**
 * 所有语言的 name 构成的字面量类型
 * 'zh-CN' | 'en-US'
 */
export type LocaleKey = LocaleOption["value"];

// 定义消息类型，确保类型安全
export type LocaleMessages = Record<string, any>;