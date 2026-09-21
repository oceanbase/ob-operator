import { getLocale } from '@umijs/max';
export const L = (zh: string, en: string) => getLocale().startsWith('zh') ? zh : en;
export const errorText = (e: unknown) => e instanceof Error ? e.message : String(e);
