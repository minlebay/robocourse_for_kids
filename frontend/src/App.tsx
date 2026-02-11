import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from './features/auth/AuthContext'
import { ThemeProvider, ThemeSelector } from './features/theme'
import { LoginPage } from './features/auth/LoginPage'
import { RegisterPage } from './features/auth/RegisterPage'
import { CatalogPage } from './features/lessons/CatalogPage'
import { ModulePage } from './features/lessons/ModulePage'
import { LessonPage } from './features/lessons/LessonPage'
import { ProgressPage } from './features/progress/ProgressPage'
import { ParentDashboardPage } from './features/parent-dashboard/ParentDashboardPage'
import './App.css'

function Layout({ children }: { children: React.ReactNode }) {
  const { user, loading, logout } = useAuth()
  if (loading) return <div className="layout">Загрузка...</div>
  return (
    <div className="layout">
      <div className="layout-center">
        <header className="header">
          <nav>
            <a href="/">Каталог</a>
            {user && (
              <>
                <a href="/progress">Мой прогресс</a>
                {user.role === 'teacher' && (
                  <a href="/dashboard">Дашборд</a>
                )}
                <button type="button" onClick={logout}>
                  Выйти ({user.name})
                </button>
              </>
            )}
            {!user && (
              <>
                <a href="/login">Вход</a>
                <a href="/register">Регистрация</a>
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

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <p>Загрузка...</p>
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

function RequireTeacher({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <p>Загрузка...</p>
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== 'teacher') return <Navigate to="/" replace />
  return <>{children}</>
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<CatalogPage />} />
      <Route path="/modules/:id" element={<ModulePage />} />
      <Route path="/lessons/:id" element={<LessonPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        path="/progress"
        element={
          <RequireAuth>
            <ProgressPage />
          </RequireAuth>
        }
      />
      <Route
        path="/dashboard"
        element={
          <RequireTeacher>
            <ParentDashboardPage />
          </RequireTeacher>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <AuthProvider>
          <Layout>
            <AppRoutes />
          </Layout>
        </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  )
}
