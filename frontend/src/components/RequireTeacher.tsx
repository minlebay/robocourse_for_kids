import { Navigate } from 'react-router-dom'
import { useAuth } from '../features/auth/AuthContext'

export function RequireTeacher({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <p>Загрузка...</p>
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== 'teacher') return <Navigate to="/" replace />
  return <>{children}</>
}
