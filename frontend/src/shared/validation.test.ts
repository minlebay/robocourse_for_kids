import { describe, it, expect } from 'vitest'
import {
  AUTH_LOGIN_MIN,
  AUTH_LOGIN_MAX,
  AUTH_PASSWORD_MIN,
  AUTH_PASSWORD_MAX,
  CHAT_MESSAGE_MAX,
  validateLogin,
  validatePassword,
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
