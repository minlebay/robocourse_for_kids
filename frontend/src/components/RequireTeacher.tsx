import { Navigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../features/auth/AuthContext'
import { hasTeacherAccess } from '../shared/roles'

export function RequireTeacher({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  const { t } = useTranslation()
  if (loading) return <p>{t('common.loading')}</p>
  if (!user) return <Navigate to="/login" replace />
  if (!hasTeacherAccess(user)) return <Navigate to="/" replace />
  return <>{children}</>
}
