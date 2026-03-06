// src/preferences/index.ts
import { reactive, watch } from "vue";

export type BuiltinThemeType = "default" | "ant" | "shadcn" | "element";

interface ThemeConfig {
  colorPrimary: string;
  builtinType: BuiltinThemeType;
}

interface Preferences {
  theme: ThemeConfig;
}

/**
 * 默认配置
 */
const defaultPreferences: Preferences = {
  theme: {
    colorPrimary: "#1677ff",
    builtinType: "default",
  },
};

/**
 * 从本地读取
 */
const storage = localStorage.getItem("app-preferences");

export const preferences = reactive<Preferences>(
  storage ? JSON.parse(storage) : defaultPreferences,
);

/**
 * 更新方法
 */
export function updatePreferences(partial: Partial<Preferences>) {
  Object.assign(preferences, partial);
}

/**
 * 持久化
 */
watch(
  preferences,
  (val) => {
    localStorage.setItem("app-preferences", JSON.stringify(val));
    applyTheme(val.theme);
  },
  { deep: true },
);

/**
 * 应用主题到 CSS 变量
 */
function applyTheme(theme: ThemeConfig) {
  document.documentElement.style.setProperty(
    "--app-primary-color",
    theme.colorPrimary,
  );
}

/**
 * 预设颜色
 */
export const COLOR_PRESETS = [
  {
    color: "hsl(212 100% 45%)",
    type: "default",
  },
  {
    color: "hsl(245 82% 67%)",
    type: "violet",
  },
  {
    color: "hsl(347 77% 60%)",
    type: "pink",
  },
  {
    color: "hsl(42 84% 61%)",
    type: "yellow",
  },
  {
    color: "hsl(231 98% 65%)",
    type: "sky-blue",
  },
  {
    color: "hsl(161 90% 43%)",
    type: "green",
  },
  {
    color: "hsl(240 5% 26%)",
    darkPrimaryColor: "hsl(0 0% 98%)",
    primaryColor: "hsl(240 5.9% 10%)",
    type: "zinc",
  },
  {
    color: "hsl(181 84% 32%)",
    type: "deep-green",
  },
  {
    color: "hsl(211 91% 39%)",
    type: "deep-blue",
  },
  {
    color: "hsl(18 89% 40%)",
    type: "orange",
  },
  {
    color: "hsl(0 75% 42%)",
    type: "rose",
  },
  {
    color: "hsl(0 0% 25%)",
    darkPrimaryColor: "hsl(0 0% 98%)",
    primaryColor: "hsl(240 5.9% 10%)",
    type: "neutral",
  },
  {
    color: "hsl(215 25% 27%)",
    darkPrimaryColor: "hsl(0 0% 98%)",
    primaryColor: "hsl(240 5.9% 10%)",
    type: "slate",
  },
  {
    color: "hsl(217 19% 27%)",
    darkPrimaryColor: "hsl(0 0% 98%)",
    primaryColor: "hsl(240 5.9% 10%)",
    type: "gray",
  },
//   {
//     color: "",
//     type: "custom",
//   },
];
