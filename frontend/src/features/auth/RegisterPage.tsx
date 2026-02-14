import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
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
    const nameErr = validateName(name)
    if (nameErr) {
      setError(nameErr)
      return
    }
    if (role === 'teacher' && !inviteCode.trim()) {
      setError('Для роли учителя нужен код приглашения')
      return
    }

    try {
      await register(login.trim(), password, name.trim(), role, role === 'teacher' ? inviteCode.trim() : undefined)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка регистрации')
    }
  }

  return (
    <div className="auth-page">
      <h1>Регистрация</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="register-login">Логин</label>
          <input
            id="register-login"
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
          <label htmlFor="register-password">Пароль</label>
          <input
            id="register-password"
            type="password"
            placeholder="Пароль"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={6}
            maxLength={72}
            autoComplete="new-password"
          />
        </div>
        <div className="form-group">
          <label htmlFor="register-name">Имя</label>
          <input
            id="register-name"
            type="text"
            placeholder="Имя"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            maxLength={200}
            autoComplete="name"
          />
        </div>
        <div className="form-group">
          <label htmlFor="register-role">Роль</label>
          <select
            id="register-role"
            value={role}
            onChange={(e) => setRole(e.target.value as 'student' | 'teacher')}
          >
            <option value="student">Ученик</option>
            <option value="teacher">Учитель / Родитель</option>
          </select>
        </div>
        {role === 'teacher' && (
          <div className="form-group">
            <label htmlFor="register-invite-code">Код приглашения (для учителя)</label>
            <input
              id="register-invite-code"
              type="text"
              placeholder="Код приглашения (для учителя)"
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
        <button type="submit">Зарегистрироваться</button>
      </form>
      <p>
        Уже есть аккаунт? <Link to="/login">Войти</Link>
      </p>
    </div>
  )
}
