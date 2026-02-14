/** Правила валидации по docs/context/auth.md */
export const AUTH_LOGIN_MIN = 3
export const AUTH_LOGIN_MAX = 50
export const AUTH_PASSWORD_MIN = 6
export const AUTH_PASSWORD_MAX = 72
export const AUTH_NAME_MAX = 200

export function validateLogin(login: string): string | null {
  const t = login.trim()
  if (t.length < AUTH_LOGIN_MIN) return `Логин не менее ${AUTH_LOGIN_MIN} символов`
  if (t.length > AUTH_LOGIN_MAX) return `Логин не более ${AUTH_LOGIN_MAX} символов`
  return null
}

export function validatePassword(password: string): string | null {
  if (password.length < AUTH_PASSWORD_MIN) return `Пароль не менее ${AUTH_PASSWORD_MIN} символов`
  if (password.length > AUTH_PASSWORD_MAX) return `Пароль не более ${AUTH_PASSWORD_MAX} символов`
  return null
}

export function validateName(name: string): string | null {
  const t = name.trim()
  if (!t) return 'Введите имя'
  if (t.length > AUTH_NAME_MAX) return `Имя не более ${AUTH_NAME_MAX} символов`
  return null
}

/** Лимит сообщения чата (docs/context/chat.md). */
export const CHAT_MESSAGE_MAX = 1000

/** Лимит текста комментария (docs/context/comments.md). */
export const COMMENT_MAX = 2000

export function validateChatMessage(text: string): string | null {
  const t = text.trim()
  if (!t) return 'Введите сообщение'
  if (t.length > CHAT_MESSAGE_MAX) return `Не более ${CHAT_MESSAGE_MAX} символов`
  return null
}
