import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../features/auth/AuthContext'
import { HeaderResumeHint } from '../features/progress'
import { ThemeSelector } from '../features/theme'
import { LanguageSelector } from '../features/i18n/LanguageSelector'
import { hasTeacherAccess, hasAdminAccess } from '../shared/roles'

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
                {hasTeacherAccess(user) && (
                  <>
                    <Link to="/my-courses">{t('nav.myCourses')}</Link>
                    <Link to="/dashboard">{t('nav.dashboard')}</Link>
                  </>
                )}
                {hasAdminAccess(user) && (
                  <Link to="/admin">{t('nav.admin')}</Link>
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
