import { ref } from 'vue';

const LOCALE_KEY = 'app-locale';

// 默认语言
const localeRef = ref<string>(
  localStorage.getItem(LOCALE_KEY) || 'zh-CN',
);

/**
 * 设置语言
 */
export function setSimpleLocale(locale: string) {
  localeRef.value = locale;
  localStorage.setItem(LOCALE_KEY, locale);
}

/**
 * 获取语言
 */
export function useSimpleLocale() {
  return {
    locale: localeRef,
    setSimpleLocale,
  };
}