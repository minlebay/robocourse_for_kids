import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { ru } from './locales/ru'
import { en } from './locales/en'

const resources = {
  ru: { translation: ru },
  en: { translation: en },
}

let defaultLng = 'ru'
try {
  if (typeof window !== 'undefined' && typeof localStorage?.getItem === 'function') {
    const saved = localStorage.getItem('learn_kids_lang')
    if (saved === 'en' || saved === 'ru') defaultLng = saved
  }
} catch {
  // localStorage not available (e.g. in tests)
}

i18n.use(initReactI18next).init({
  resources,
  lng: defaultLng,
  fallbackLng: 'ru',
  defaultNS: 'translation',
  interpolation: {
    escapeValue: false,
  },
  react: {
    useSuspense: false,
  },
})

i18n.on('languageChanged', (lng) => {
  if (typeof document?.documentElement !== 'undefined') {
    document.documentElement.lang = lng === 'en' ? 'en' : 'ru'
  }
  try {
    localStorage.setItem('learn_kids_lang', lng)
  } catch {
    // ignore
  }
})

// set initial html lang
if (typeof document?.documentElement !== 'undefined') {
  document.documentElement.lang = i18n.language === 'en' ? 'en' : 'ru'
}

export default i18n
