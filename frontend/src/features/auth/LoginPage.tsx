import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from './AuthContext'
import { consumeReturnUrl } from '../../shared/api'
import { validateLogin, validatePassword } from '../../shared/validation'

export function LoginPage() {
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const { login: doLogin } = useAuth()
  const navigate = useNavigate()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    const loginErr = validateLogin(login)
    if (loginErr) {
      setError(loginErr)
      return
    }
    const passwordErr = validatePassword(password)
    if (passwordErr) {
      setError(passwordErr)
      return
    }

    try {
      await doLogin(login.trim(), password)
      const returnUrl = consumeReturnUrl()
      navigate(returnUrl ?? '/', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка входа')
    }
  }

  return (
    <div className="auth-page">
      <h1>Вход</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="login-username">Логин</label>
          <input
            id="login-username"
            type="text"
            placeholder="Логин"
            value={login}
            onChange={(e) => setLogin(e.target.value)}
            required
            minLength={3}
            maxLength={50}
            autoComplete="username"
          />
        </div>
        <div className="form-group">
          <label htmlFor="login-password">Пароль</label>
          <input
            id="login-password"
            type="password"
            placeholder="Пароль"
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
        <button type="submit">Войти</button>
      </form>
      <p>
        Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
      </p>
    </div>
  )
}
