import { describe, it, expect } from 'vitest'
import {
  AUTH_LOGIN_MIN,
  AUTH_LOGIN_MAX,
  AUTH_PASSWORD_MIN,
  AUTH_PASSWORD_MAX,
  AUTH_NAME_MAX,
  CHAT_MESSAGE_MAX,
  COMMENT_MAX,
  validateLogin,
  validatePassword,
  validateName,
  validateChatMessage,
} from './validation'

describe('validateLogin', () => {
  it('returns error when login too short', () => {
    expect(validateLogin('ab')).toBe(`Логин не менее ${AUTH_LOGIN_MIN} символов`)
    expect(validateLogin('')).toBe(`Логин не менее ${AUTH_LOGIN_MIN} символов`)
  })

  it('returns error when login too long', () => {
    const long = 'a'.repeat(AUTH_LOGIN_MAX + 1)
    expect(validateLogin(long)).toBe(`Логин не более ${AUTH_LOGIN_MAX} символов`)
  })

  it('returns null for valid login', () => {
    expect(validateLogin('abc')).toBeNull()
    expect(validateLogin('  valid  ')).toBeNull()
    expect(validateLogin('a'.repeat(AUTH_LOGIN_MAX))).toBeNull()
  })
})

describe('validatePassword', () => {
  it('returns error when password too short', () => {
    expect(validatePassword('12345')).toBe(`Пароль не менее ${AUTH_PASSWORD_MIN} символов`)
  })

  it('returns error when password too long', () => {
    const long = 'a'.repeat(AUTH_PASSWORD_MAX + 1)
    expect(validatePassword(long)).toBe(`Пароль не более ${AUTH_PASSWORD_MAX} символов`)
  })

  it('returns null for valid password', () => {
    expect(validatePassword('123456')).toBeNull()
    expect(validatePassword('a'.repeat(AUTH_PASSWORD_MAX))).toBeNull()
  })
})

describe('validateName', () => {
  it('returns error for empty or whitespace', () => {
    expect(validateName('')).toBe('Введите имя')
    expect(validateName('   ')).toBe('Введите имя')
  })

  it('returns error when name too long', () => {
    const long = 'а'.repeat(AUTH_NAME_MAX + 1)
    expect(validateName(long)).toBe(`Имя не более ${AUTH_NAME_MAX} символов`)
  })

  it('returns null for valid name', () => {
    expect(validateName('Иван')).toBeNull()
    expect(validateName('  trim  ')).toBeNull()
    expect(validateName('а'.repeat(AUTH_NAME_MAX))).toBeNull()
  })
})

describe('validateChatMessage', () => {
  it('returns error for empty or whitespace', () => {
    expect(validateChatMessage('')).toBe('Введите сообщение')
    expect(validateChatMessage('   ')).toBe('Введите сообщение')
  })

  it('returns error when over limit', () => {
    const long = 'a'.repeat(CHAT_MESSAGE_MAX + 1)
    expect(validateChatMessage(long)).toBe(`Не более ${CHAT_MESSAGE_MAX} символов`)
  })

  it('returns null for valid message', () => {
    expect(validateChatMessage('hello')).toBeNull()
    expect(validateChatMessage('  trimmed  ')).toBeNull()
    expect(validateChatMessage('a'.repeat(CHAT_MESSAGE_MAX))).toBeNull()
  })
})

describe('COMMENT_MAX', () => {
  it('is 2000', () => {
    expect(COMMENT_MAX).toBe(2000)
  })
})
