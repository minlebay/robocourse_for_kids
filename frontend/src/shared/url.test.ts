import { describe, it, expect } from 'vitest'
import { isSafeHttpUrl, getSafeUrl } from './url'

describe('isSafeHttpUrl', () => {
  it('allows https URLs', () => {
    expect(isSafeHttpUrl('https://example.com')).toBe(true)
    expect(isSafeHttpUrl('https://example.com/path')).toBe(true)
  })

  it('allows http URLs', () => {
    expect(isSafeHttpUrl('http://example.com')).toBe(true)
  })

  it('rejects javascript: URLs', () => {
    expect(isSafeHttpUrl('javascript:alert(1)')).toBe(false)
  })

  it('rejects data: URLs', () => {
    expect(isSafeHttpUrl('data:text/html,<script>')).toBe(false)
  })

  it('rejects non-http(s) protocols', () => {
    expect(isSafeHttpUrl('file:///etc/passwd')).toBe(false)
  })
})

describe('getSafeUrl', () => {
  it('returns URL for safe http', () => {
    expect(getSafeUrl('https://img.example.com/x.png')).toBe('https://img.example.com/x.png')
  })

  it('returns null for javascript:', () => {
    expect(getSafeUrl('javascript:void(0)')).toBeNull()
  })

  it('returns null for empty string', () => {
    expect(getSafeUrl('')).toBeNull()
  })

  it('trims and validates', () => {
    expect(getSafeUrl('  https://a.com  ')).toBe('https://a.com')
  })
})
