import { Link } from 'react-router-dom'
import { useAuth } from '../features/auth/AuthContext'
import { HeaderResumeHint } from '../features/progress'
import { ThemeSelector } from '../features/theme'

export function Layout({ children }: { children: React.ReactNode }) {
  const { user, loading, logout } = useAuth()
  if (loading) return <div className="layout">Загрузка...</div>
  return (
    <div className="layout">
      <div className="layout-center">
        <header className="header">
          <nav>
            <Link to="/">Каталог</Link>
            {user && (
              <>
                <HeaderResumeHint />
                <Link to="/progress">Мой прогресс</Link>
                {user.role === 'teacher' && (
                  <Link to="/dashboard">Дашборд</Link>
                )}
                <button type="button" onClick={logout}>
                  Выйти ({user.name})
                </button>
              </>
            )}
            {!user && (
              <>
                <Link to="/login">Вход</Link>
                <Link to="/register">Регистрация</Link>
              </>
            )}
            <ThemeSelector />
          </nav>
        </header>
        <main className="main">{children}</main>
      </div>
    </div>
  )
}
