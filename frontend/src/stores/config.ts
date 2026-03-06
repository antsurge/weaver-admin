import { defineStore } from "pinia";
import { reactive } from "vue";
import { STORE_CONFIG } from "@/stores/constant/cacheKey";
import type { Lang } from "@/stores/interface";
import { LOCALE_ARRAY, DEFAULT_LOCALE } from "@/locales/constant";

export const useConfigStore = defineStore(
  "config",
  () => {
    // 语言包
    const lang: Lang = reactive({
      defaultLang: DEFAULT_LOCALE.value,
      fallbackLang: DEFAULT_LOCALE.value,
      langArray: LOCALE_ARRAY,
    });

    function setLang(val: string) {
      lang.defaultLang = val;
    }

    return { lang, setLang };
  },
  {
    persist: {
      key: STORE_CONFIG,
    },
  },
);
