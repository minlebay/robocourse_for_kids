import { Suspense, lazy } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './features/auth/AuthProvider'
import { ThemeProvider } from './features/theme'
import { ErrorBoundary } from './shared/ErrorBoundary'
import { Layout, RequireAuth, RequireTeacher } from './components'
import { RequireAdmin } from './components/RequireAdmin'
import { RequirePasswordChange } from './components/RequirePasswordChange'
import './App.css'

const LoginPage = lazy(() => import('./features/auth/LoginPage').then((m) => ({ default: m.LoginPage })))
const RegisterPage = lazy(() => import('./features/auth/RegisterPage').then((m) => ({ default: m.RegisterPage })))
const ChangePasswordPage = lazy(() => import('./features/auth/ChangePasswordPage').then((m) => ({ default: m.ChangePasswordPage })))
const CatalogPage = lazy(() => import('./features/lessons/CatalogPage').then((m) => ({ default: m.CatalogPage })))
const ModulePage = lazy(() => import('./features/lessons/ModulePage').then((m) => ({ default: m.ModulePage })))
const LessonPage = lazy(() => import('./features/lessons/LessonPage').then((m) => ({ default: m.LessonPage })))
const ProgressPage = lazy(() => import('./features/progress').then((m) => ({ default: m.ProgressPage })))
const ParentDashboardPage = lazy(() => import('./features/parent-dashboard/ParentDashboardPage').then((m) => ({ default: m.ParentDashboardPage })))
const AdminPage = lazy(() => import('./features/admin/AdminPage').then((m) => ({ default: m.AdminPage })))

function LoadingFallback() {
  return (
    <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
      Загрузка...
    </div>
  )
}

function AppRoutes() {
  return (
    <RequirePasswordChange>
      <Suspense fallback={<LoadingFallback />}>
        <Routes>
          <Route path="/" element={<CatalogPage />} />
          <Route path="/modules/:id" element={<ModulePage />} />
          <Route path="/lessons/:id" element={<LessonPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route
            path="/change-password"
            element={
              <RequireAuth>
                <ChangePasswordPage />
              </RequireAuth>
            }
          />
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
          <Route
            path="/admin"
            element={
              <RequireAdmin>
                <AdminPage />
              </RequireAdmin>
            }
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
    </RequirePasswordChange>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ThemeProvider>
          <Layout>
            <ErrorBoundary>
              <AppRoutes />
            </ErrorBoundary>
          </Layout>
        </ThemeProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}
