import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { AdminCreateUserRequest, User } from '../../shared/types'

interface CreateUserModalProps {
  onClose: () => void
  onCreate: (data: AdminCreateUserRequest) => Promise<{ user: User; temp_password: string }>
}

export function CreateUserModal({ onClose, onCreate }: CreateUserModalProps) {
  const { t } = useTranslation()
  const [login, setLogin] = useState('')
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<'student' | 'teacher'>('student')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (password.length < 6) {
      setError(t('validation.password_min', { min: 6 }))
      return
    }

    const data: AdminCreateUserRequest = {
      login: login.trim(),
      name: name.trim(),
      password,
      role,
    }
    if (email.trim()) {
      data.email = email.trim()
    }

    setLoading(true)
    try {
      await onCreate(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('errors.somethingWrong'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="temp-modal-overlay" role="dialog" aria-modal="true" aria-labelledby="create-user-title">
      <div className="temp-modal temp-modal--wide">
        <h2 id="create-user-title" className="temp-modal-title">{t('admin.createTitle')}</h2>
        <form onSubmit={handleSubmit} className="admin-form-body">
          <div className="form-group">
            <label htmlFor="cu-login">{t('admin.loginLabel')}</label>
            <input
              id="cu-login"
              type="text"
              value={login}
              onChange={(e) => setLogin(e.target.value)}
              required
              minLength={3}
              maxLength={50}
              autoComplete="username"
              className="admin-form-control"
            />
          </div>
          <div className="form-group">
            <label htmlFor="cu-name">{t('admin.nameLabel')}</label>
            <input
              id="cu-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              maxLength={100}
              className="admin-form-control"
            />
          </div>
          <div className="form-group">
            <label htmlFor="cu-password">{t('admin.passwordLabel')}</label>
            <input
              id="cu-password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
              maxLength={72}
              autoComplete="new-password"
              className="admin-form-control"
            />
          </div>
          <div className="form-group">
            <label htmlFor="cu-role">{t('admin.roleLabel')}</label>
            <select
              id="cu-role"
              value={role}
              onChange={(e) => setRole(e.target.value as 'student' | 'teacher')}
              className="admin-form-control"
            >
              <option value="student">{t('admin.roleStudent')}</option>
              <option value="teacher">{t('admin.roleTeacher')}</option>
            </select>
          </div>
          <div className="form-group">
            <label htmlFor="cu-email">{t('admin.emailLabel')}</label>
            <input
              id="cu-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              maxLength={255}
              className="admin-form-control"
            />
          </div>
          {error && (
            <p className="error" role="alert" style={{ margin: 0 }}>
              {error}
            </p>
          )}
          <div className="temp-modal-actions">
            <button type="submit" className="button-primary" disabled={loading}>
              {t('admin.createSubmit')}
            </button>
            <button type="button" className="button-secondary" onClick={onClose} disabled={loading}>
              {t('common.cancel')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
