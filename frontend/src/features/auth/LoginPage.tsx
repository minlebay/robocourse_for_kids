import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from './AuthContext'
import { consumeReturnUrl } from '../../shared/api'
import { validateLogin, validatePassword } from '../../shared/validation'

export function LoginPage() {
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const { login: doLogin } = useAuth()
  const navigate = useNavigate()
  const { t } = useTranslation()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    const loginErr = validateLogin(login)
    if (loginErr) {
      setError(t(loginErr.key, loginErr.params as Record<string, string | number>))
      return
    }
    const passwordErr = validatePassword(password)
    if (passwordErr) {
      setError(t(passwordErr.key, passwordErr.params as Record<string, string | number>))
      return
    }

    try {
      await doLogin(login.trim(), password)
      const returnUrl = consumeReturnUrl()
      navigate(returnUrl ?? '/', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : t('auth.loginError'))
    }
  }

  return (
    <div className="auth-page">
      <h1>{t('auth.title')}</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="login-username">{t('auth.login')}</label>
          <input
            id="login-username"
            type="text"
            placeholder={t('auth.login')}
            value={login}
            onChange={(e) => setLogin(e.target.value)}
            required
            minLength={3}
            maxLength={50}
            autoComplete="username"
          />
        </div>
        <div className="form-group">
          <label htmlFor="login-password">{t('auth.password')}</label>
          <input
            id="login-password"
            type="password"
            placeholder={t('auth.password')}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={6}
            maxLength={72}
            autoComplete="current-password"
          />
        </div>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <button type="submit">{t('auth.enter')}</button>
      </form>
      <p>
        {t('auth.noAccount')} <Link to="/register">{t('auth.register')}</Link>
      </p>
    </div>
  )
}
