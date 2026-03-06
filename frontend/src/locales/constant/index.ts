// src/locales/constants.ts
import type { LocaleOption } from '@/locales/interface';

/**
 * 所有语言常量
 */
export const LOCALES: Record<string, LocaleOption> = {
  ZH_CN: { value: 'zh-CN', label: '中文简体' },
  EN_US: { value: 'en-US', label: 'English' },
} as const;

/**
 * 默认语言
 */
export const DEFAULT_LOCALE: LocaleOption = LOCALES.ZH_CN!;

/**
 * 下拉选择器语言列表
 */
export const LOCALE_ARRAY: LocaleOption[] = Object.values(LOCALES);