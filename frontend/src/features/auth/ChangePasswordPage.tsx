import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from './AuthContext'
import { auth as authApi } from '../../shared/api'

export function ChangePasswordPage() {
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [loading, setLoading] = useState(false)

  const { user, setUser } = useAuth()
  const navigate = useNavigate()
  const { t } = useTranslation()

  const isForced = user?.must_change_password === true

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (newPassword.length < 6 || newPassword.length > 72) {
      setError(t('changePassword.errorWeak'))
      return
    }
    if (newPassword !== confirmPassword) {
      setError(t('changePassword.errorMismatch'))
      return
    }

    setLoading(true)
    try {
      await authApi.changePassword(currentPassword, newPassword)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('changePassword.errorCurrent'))
      setLoading(false)
      return
    }

    // Пароль изменён — сбрасываем флаг оптимистично, чтобы избежать зацикленного
    // редиректа на /change-password если последующий me() упадёт.
    setUser(user ? { ...user, must_change_password: false } : null)
    try {
      const updatedUser = await authApi.me()
      setUser(updatedUser)
    } catch {
      // Игнорируем: флаг уже сброшен оптимистично выше
    }
    setSuccess(true)
    setLoading(false)
    setTimeout(() => {
      navigate('/', { replace: true })
    }, 1500)
  }

  return (
    <div className="auth-page">
      <h1>{isForced ? t('changePassword.forcedTitle') : t('changePassword.title')}</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="current-password">{t('changePassword.currentPassword')}</label>
          <input
            id="current-password"
            type="password"
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            required
            autoComplete="current-password"
            style={{ width: '100%' }}
          />
        </div>
        <div className="form-group">
          <label htmlFor="new-password">{t('changePassword.newPassword')}</label>
          <input
            id="new-password"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            required
            minLength={6}
            maxLength={72}
            autoComplete="new-password"
            style={{ width: '100%' }}
          />
        </div>
        <div className="form-group">
          <label htmlFor="confirm-password">{t('changePassword.confirmPassword')}</label>
          <input
            id="confirm-password"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            minLength={6}
            maxLength={72}
            autoComplete="new-password"
            style={{ width: '100%' }}
          />
        </div>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        {success && (
          <p style={{ color: 'var(--success)' }} role="status">
            {t('changePassword.success')}
          </p>
        )}
        <button type="submit" disabled={loading}>
          {t('changePassword.submit')}
        </button>
      </form>
      {!isForced && (
        <p>
          <button
            type="button"
            onClick={() => navigate(-1)}
            style={{
              background: 'transparent',
              border: 'none',
              color: 'var(--accent)',
              cursor: 'pointer',
              fontWeight: 600,
              padding: 0,
              font: 'inherit',
            }}
          >
            {t('common.back')}
          </button>
        </p>
      )}
    </div>
  )
}
