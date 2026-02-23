import { Navigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../features/auth/AuthContext'

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  const { t } = useTranslation()
  if (loading) return <p>{t('common.loading')}</p>
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}
