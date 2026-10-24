import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'

export type Locale = 'zh-CN' | 'en-US' | 'ja-JP'

export const LOCALES: { value: Locale; label: string }[] = [
  { value: 'zh-CN', label: '中文' },
  { value: 'en-US', label: 'English' },
  { value: 'ja-JP', label: '日本語' }
]

export type MessageSchema = typeof zhCN

// 语言包按需加载（2026-10-22 首屏优化）：三语言全量静态引入曾让 locales
// chunk 达 316KB，而首屏只用一种语言。默认语言（zh-CN）同步注入保证首屏
// 无闪烁；其余语言动态 import，切换时按需拉取。
const messageLoaders: Record<Locale, () => Promise<{ default: MessageSchema }>> = {
  'zh-CN': async () => ({ default: zhCN }),
  'en-US': () => import('./en-US'),
  'ja-JP': () => import('./ja-JP')
}

const loadedLocales = new Set<Locale>(['zh-CN'])

/** 确保某语言包已注入（切换语言前调用；已加载则直接返回） */
export async function ensureLocale(l: Locale): Promise<void> {
  if (loadedLocales.has(l)) return
  const mod = await messageLoaders[l]()
  i18n.global.setLocaleMessage(l, mod.default)
  loadedLocales.add(l)
}

const savedLocale = (localStorage.getItem('cloudpivot_locale') as Locale) || 'zh-CN'

const i18n = createI18n<{ message: MessageSchema }, Locale, false>({
  legacy: false,
  // 初始一律中文：非默认语言的用户在 bootstrap 里经 initLocale() 无缝切换，
  // 避免首屏先渲染默认语言再跳变
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    // en-US / ja-JP 运行时由 ensureLocale → setLocaleMessage 注入真实文案，
    // 这里给空占位只为满足 createI18n 的 Record<Locale, Message> 泛型约束
    'en-US': {} as MessageSchema,
    'ja-JP': {} as MessageSchema
  }
})

/** bootstrap 调用：加载用户保存的语言包并切换（默认语言同步可用，无闪烁） */
export async function initLocale(): Promise<void> {
  if (savedLocale !== 'zh-CN') {
    await ensureLocale(savedLocale)
    i18n.global.locale.value = savedLocale
  }
}

export default i18n
