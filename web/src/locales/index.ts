import { createI18n } from 'vue-i18n'
import en from './en'
import zh from './zh'

export type Locale = 'en' | 'zh'

// 获取存储的语言或浏览器语言
function getDefaultLocale(): Locale {
  const stored = localStorage.getItem('locale')
  if (stored && (stored === 'en' || stored === 'zh')) {
    return stored as Locale
  }
  
  // 检测浏览器语言
  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) {
    return 'zh'
  }
  return 'en'
}

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: getDefaultLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    zh,
  },
})

export default i18n

export function setLocale(locale: Locale): void {
  i18n.global.locale.value = locale
  localStorage.setItem('locale', locale)
  document.documentElement.setAttribute('lang', locale)
}

export function getLocale(): Locale {
  return i18n.global.locale.value as Locale
}
