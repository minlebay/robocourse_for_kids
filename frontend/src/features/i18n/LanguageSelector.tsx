import { useTranslation } from 'react-i18next'

export function LanguageSelector() {
  const { i18n } = useTranslation()

  return (
    <span className="language-selector">
      <button
        type="button"
        className={i18n.language === 'ru' ? 'active' : ''}
        onClick={() => i18n.changeLanguage('ru')}
        aria-pressed={i18n.language === 'ru'}
        aria-label="Русский"
      >
        RU
      </button>
      <button
        type="button"
        className={i18n.language === 'en' ? 'active' : ''}
        onClick={() => i18n.changeLanguage('en')}
        aria-pressed={i18n.language === 'en'}
        aria-label="English"
      >
        EN
      </button>
    </span>
  )
}
