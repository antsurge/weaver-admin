import type { LocaleOption } from "@/locales/interface";

export interface Lang {
  defaultLang: string;
  fallbackLang: string;
  langArray: LocaleOption[];
}
