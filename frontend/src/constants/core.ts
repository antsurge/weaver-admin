/**
 * @zh_CN 登录页面 url 地址
 */
export const LOGIN_PATH = "/auth/login";

export interface LanguageOption {
  label: string;
  value: "en-US" | "zh-CN";
}

/**
 * Supported languages
 */
export const SUPPORT_LANGUAGES: LanguageOption[] = [
  {
    label: "简体中文",
    value: "zh-CN",
  },
  {
    label: "English",
    value: "en-US",
  },
];

/** layout content 组件的高度 */
export const CSS_VARIABLE_LAYOUT_CONTENT_HEIGHT = `--vben-content-height`;
/** layout content 组件的宽度 */
export const CSS_VARIABLE_LAYOUT_CONTENT_WIDTH = `--vben-content-width`;
/** layout header 组件的高度 */
export const CSS_VARIABLE_LAYOUT_HEADER_HEIGHT = `--vben-header-height`;
/** layout footer 组件的高度 */
export const CSS_VARIABLE_LAYOUT_FOOTER_HEIGHT = `--vben-footer-height`;

/** 内容区域的组件ID */
export const ELEMENT_ID_MAIN_CONTENT = `__vben_main_content`;

/**
 * @zh_CN 默认命名空间
 */
export const DEFAULT_NAMESPACE = 'vben';