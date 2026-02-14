import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './features/auth/AuthContext'
import { ThemeProvider } from './features/theme'
import { ErrorBoundary } from './shared/ErrorBoundary'
import { Layout, RequireAuth, RequireTeacher } from './components'
import { LoginPage } from './features/auth/LoginPage'
import { RegisterPage } from './features/auth/RegisterPage'
import { CatalogPage } from './features/lessons/CatalogPage'
import { ModulePage } from './features/lessons/ModulePage'
import { LessonPage } from './features/lessons/LessonPage'
import { ProgressPage } from './features/progress'
import { ParentDashboardPage } from './features/parent-dashboard/ParentDashboardPage'
import './App.css'

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
