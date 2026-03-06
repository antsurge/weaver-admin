import type { App } from "vue";
import type { I18n, Composer } from "vue-i18n";
import { createI18n } from "vue-i18n";
import { useConfigStore } from "@/stores/config";
import type { LocaleMessages } from "@/locales/interface";

let i18n: ReturnType<typeof createI18n>;

/**
 * 动态加载所有语言文件
 * 假设目录结构：src/locales/langs/{locale}/{fileName}.ts
 */
async function loadAllMessages(): Promise<Record<string, LocaleMessages>> {
  // 1. 使用 import.meta.glob 加载 langs 目录下所有 .ts 文件
  // eager: false 表示按需加载（异步），如果文件很少也可以设为 true 同步加载
  const modules = import.meta.glob<{ default: LocaleMessages }>(
    "./langs/*/*.ts",
    { eager: false },
  );

  const messages: Record<string, LocaleMessages> = {};

  // 2. 遍历所有匹配的文件
  for (const path in modules) {
    // 解析路径：./langs/zh-CN/common.ts -> locale='zh-CN', fileName='common'
    const match = path.match(/langs\/([^/]+)\/([^/]+)\.ts$/);
    if (!match) continue;

    const locale = match[1] as string;
    const fileName = match[2] as string;

    // 初始化该语言的容器
    if (!messages[locale]) {
      messages[locale] = {};
    }

    try {
      // 动态导入文件
      const mod = await modules[path]!();

      // 【关键点】将文件内容赋值给 fileName 作为 key
      // 这样结构就变成了：messages['zh-CN']['common'] = { ... }
      if (mod.default) {
        messages[locale][fileName] = mod.default;
      }
    } catch (err) {
      console.error(`Failed to load locale file: ${path}`, err);
    }
  }

  return messages;
}

export async function loadLang(app: App) {
  const config = useConfigStore();
  const locale = config.lang.defaultLang;
  // 加载所有的消息
  const messages = await loadAllMessages();

  i18n = createI18n({
    locale: locale,
    legacy: false,
    globalInjection: true,
    fallbackLocale: config.lang.fallbackLang,
    messages,
  });

  app.use(i18n as I18n);

  return i18n;
}

/**
 * 翻译函数
 */
export const $t = (...args: Parameters<Composer["t"]>) => {
  return (i18n.global as Composer).t(...args);
};

/**
 * 判断key是否存在
 */
export const $te = (...args: Parameters<Composer["te"]>) => {
  return (i18n.global as Composer).te(...args);
};

/**
 * 切换当前语言
 */
export async function setI18nLanguage(lang: string) {
  const config = useConfigStore();

  // 1. 如果当前语言就是目标语言，直接返回
  if (lang === (i18n.global as Composer).locale.value) return;

  // 2. 保存到 Pinia
  config.setLang(lang);

  // 3. 更新 i18n 当前语言
  (i18n.global as Composer).locale.value = lang;

  // 4. 如果消息里还没有这个语言，动态加载语言包
  if (!(i18n.global.getLocaleMessage(lang) && Object.keys(i18n.global.getLocaleMessage(lang)).length)) {
    const messages = await loadAllMessages();
    if (messages[lang]) {
      i18n.global.setLocaleMessage(lang, messages[lang]);
    }
  }
}