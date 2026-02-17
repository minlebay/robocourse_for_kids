import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../features/auth/AuthContext'
import { HeaderResumeHint } from '../features/progress'
import { ThemeSelector } from '../features/theme'
import { LanguageSelector } from '../features/i18n/LanguageSelector'

export function Layout({ children }: { children: React.ReactNode }) {
  const { user, loading, logout } = useAuth()
  const { t } = useTranslation()

  if (loading) return <div className="layout">{t('common.loading')}</div>
  return (
    <div className="layout">
      <div className="layout-center">
        <header className="header">
          <nav>
            <Link to="/">{t('nav.catalog')}</Link>
            {user && (
              <>
                <HeaderResumeHint />
                <Link to="/progress">{t('nav.progress')}</Link>
                {user.role === 'teacher' && (
                  <Link to="/dashboard">{t('nav.dashboard')}</Link>
                )}
                <button type="button" onClick={logout}>
                  {t('nav.logout')} ({user.name})
                </button>
              </>
            )}
            {!user && (
              <>
                <Link to="/login">{t('nav.login')}</Link>
                <Link to="/register">{t('nav.register')}</Link>
              </>
            )}
            <span className="header-actions">
              <LanguageSelector />
              <ThemeSelector />
            </span>
          </nav>
        </header>
        <main className="main">{children}</main>
      </div>
    </div>
  )
}
