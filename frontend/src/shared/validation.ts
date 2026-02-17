/** Правила валидации по docs/context/auth.md. Возвращают ключи i18n и параметры для интерполяции. */
export const AUTH_LOGIN_MIN = 3
export const AUTH_LOGIN_MAX = 50
export const AUTH_PASSWORD_MIN = 6
export const AUTH_PASSWORD_MAX = 72
export const AUTH_NAME_MAX = 200

export type ValidationError = { key: string; params?: Record<string, number> }

export function validateLogin(login: string): ValidationError | null {
  const t = login.trim()
  if (t.length < AUTH_LOGIN_MIN) return { key: 'validation.login_min', params: { min: AUTH_LOGIN_MIN } }
  if (t.length > AUTH_LOGIN_MAX) return { key: 'validation.login_max', params: { max: AUTH_LOGIN_MAX } }
  return null
}

export function validatePassword(password: string): ValidationError | null {
  if (password.length < AUTH_PASSWORD_MIN)
    return { key: 'validation.password_min', params: { min: AUTH_PASSWORD_MIN } }
  if (password.length > AUTH_PASSWORD_MAX)
    return { key: 'validation.password_max', params: { max: AUTH_PASSWORD_MAX } }
  return null
}

export function validateName(name: string): ValidationError | null {
  const t = name.trim()
  if (!t) return { key: 'validation.name_required' }
  if (t.length > AUTH_NAME_MAX) return { key: 'validation.name_max', params: { max: AUTH_NAME_MAX } }
  return null
}

/** Лимит сообщения чата (docs/context/chat.md). */
export const CHAT_MESSAGE_MAX = 1000

/** Лимит текста комментария (docs/context/comments.md). */
export const COMMENT_MAX = 2000

export function validateChatMessage(text: string): ValidationError | null {
  const t = text.trim()
  if (!t) return { key: 'validation.message_required' }
  if (t.length > CHAT_MESSAGE_MAX)
    return { key: 'validation.message_max', params: { max: CHAT_MESSAGE_MAX } }
  return null
}
