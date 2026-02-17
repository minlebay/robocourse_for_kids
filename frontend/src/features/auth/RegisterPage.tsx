import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from './AuthContext'
import { validateLogin, validatePassword, validateName } from '../../shared/validation'

export function RegisterPage() {
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [role, setRole] = useState<'student' | 'teacher'>('student')
  const [inviteCode, setInviteCode] = useState('')
  const [error, setError] = useState('')
  const { register } = useAuth()
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
    const nameErr = validateName(name)
    if (nameErr) {
      setError(t(nameErr.key, nameErr.params as Record<string, string | number>))
      return
    }
    if (role === 'teacher' && !inviteCode.trim()) {
      setError(t('auth.inviteCodeRequired'))
      return
    }

    try {
      await register(login.trim(), password, name.trim(), role, role === 'teacher' ? inviteCode.trim() : undefined)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : t('auth.registerError'))
    }
  }

  return (
    <div className="auth-page">
      <h1>{t('auth.registerTitle')}</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="register-login">{t('auth.login')}</label>
          <input
            id="register-login"
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
          <label htmlFor="register-password">{t('auth.password')}</label>
          <input
            id="register-password"
            type="password"
            placeholder={t('auth.password')}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={6}
            maxLength={72}
            autoComplete="new-password"
          />
        </div>
        <div className="form-group">
          <label htmlFor="register-name">{t('auth.name')}</label>
          <input
            id="register-name"
            type="text"
            placeholder={t('auth.name')}
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            maxLength={200}
            autoComplete="name"
          />
        </div>
        <div className="form-group">
          <label htmlFor="register-role">{t('auth.role')}</label>
          <select
            id="register-role"
            value={role}
            onChange={(e) => setRole(e.target.value as 'student' | 'teacher')}
          >
            <option value="student">{t('auth.roleStudent')}</option>
            <option value="teacher">{t('auth.roleTeacher')}</option>
          </select>
        </div>
        {role === 'teacher' && (
          <div className="form-group">
            <label htmlFor="register-invite-code">{t('auth.inviteCode')}</label>
            <input
              id="register-invite-code"
              type="text"
              placeholder={t('auth.inviteCode')}
              value={inviteCode}
              onChange={(e) => setInviteCode(e.target.value)}
              required
            />
          </div>
        )}
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <button type="submit">{t('auth.register')}</button>
      </form>
      <p>
        {t('auth.hasAccount')} <Link to="/login">{t('auth.enter')}</Link>
      </p>
    </div>
  )
}
