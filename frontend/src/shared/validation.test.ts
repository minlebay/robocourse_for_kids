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
    expect(validateLogin('ab')).toEqual({ key: 'validation.login_min', params: { min: AUTH_LOGIN_MIN } })
    expect(validateLogin('')).toEqual({ key: 'validation.login_min', params: { min: AUTH_LOGIN_MIN } })
  })

  it('returns error when login too long', () => {
    const long = 'a'.repeat(AUTH_LOGIN_MAX + 1)
    expect(validateLogin(long)).toEqual({ key: 'validation.login_max', params: { max: AUTH_LOGIN_MAX } })
  })

  it('returns null for valid login', () => {
    expect(validateLogin('abc')).toBeNull()
    expect(validateLogin('  valid  ')).toBeNull()
    expect(validateLogin('a'.repeat(AUTH_LOGIN_MAX))).toBeNull()
  })
})

describe('validatePassword', () => {
  it('returns error when password too short', () => {
    expect(validatePassword('12345')).toEqual({ key: 'validation.password_min', params: { min: AUTH_PASSWORD_MIN } })
  })

  it('returns error when password too long', () => {
    const long = 'a'.repeat(AUTH_PASSWORD_MAX + 1)
    expect(validatePassword(long)).toEqual({ key: 'validation.password_max', params: { max: AUTH_PASSWORD_MAX } })
  })

  it('returns null for valid password', () => {
    expect(validatePassword('123456')).toBeNull()
    expect(validatePassword('a'.repeat(AUTH_PASSWORD_MAX))).toBeNull()
  })
})

describe('validateName', () => {
  it('returns error for empty or whitespace', () => {
    expect(validateName('')).toEqual({ key: 'validation.name_required' })
    expect(validateName('   ')).toEqual({ key: 'validation.name_required' })
  })

  it('returns error when name too long', () => {
    const long = 'а'.repeat(AUTH_NAME_MAX + 1)
    expect(validateName(long)).toEqual({ key: 'validation.name_max', params: { max: AUTH_NAME_MAX } })
  })

  it('returns null for valid name', () => {
    expect(validateName('Иван')).toBeNull()
    expect(validateName('  trim  ')).toBeNull()
    expect(validateName('а'.repeat(AUTH_NAME_MAX))).toBeNull()
  })
})

describe('validateChatMessage', () => {
  it('returns error for empty or whitespace', () => {
    expect(validateChatMessage('')).toEqual({ key: 'validation.message_required' })
    expect(validateChatMessage('   ')).toEqual({ key: 'validation.message_required' })
  })

  it('returns error when over limit', () => {
    const long = 'a'.repeat(CHAT_MESSAGE_MAX + 1)
    expect(validateChatMessage(long)).toEqual({ key: 'validation.message_max', params: { max: CHAT_MESSAGE_MAX } })
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
