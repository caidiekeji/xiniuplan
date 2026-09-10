import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'
import jaJP from './ja-JP'

export type Locale = 'zh-CN' | 'en-US' | 'ja-JP'

export const LOCALES: { value: Locale; label: string }[] = [
  { value: 'zh-CN', label: '中文' },
  { value: 'en-US', label: 'English' },
  { value: 'ja-JP', label: '日本語' }
]

export type MessageSchema = typeof zhCN

const savedLocale = (localStorage.getItem('cloudpivot_locale') as Locale) || 'zh-CN'

const i18n = createI18n<{ message: MessageSchema }, Locale, false>({
  legacy: false,
  locale: savedLocale,
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
    'ja-JP': jaJP
  }
})

export default i18n
