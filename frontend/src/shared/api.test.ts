import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { api, apiVoid, consumeReturnUrl } from './api'

const RETURN_URL_KEY = 'learn_kids_return_url'

describe('consumeReturnUrl', () => {
  const originalSessionStorage = window.sessionStorage

  beforeEach(() => {
    const store: Record<string, string> = {}
    Object.defineProperty(window, 'sessionStorage', {
      value: {
        getItem: (key: string) => store[key] ?? null,
        setItem: (key: string, value: string) => { store[key] = value },
        removeItem: (key: string) => { delete store[key] },
      },
      writable: true,
    })
  })

  afterEach(() => {
    Object.defineProperty(window, 'sessionStorage', {
      value: originalSessionStorage,
      writable: true,
    })
  })

  it('returns null when sessionStorage is empty', () => {
    expect(consumeReturnUrl()).toBeNull()
  })

  it('returns and consumes url for safe path /', () => {
    sessionStorage.setItem(RETURN_URL_KEY, '/')
    expect(consumeReturnUrl()).toBe('/')
    expect(sessionStorage.getItem(RETURN_URL_KEY)).toBeNull()
  })

  it('returns and consumes url for safe path /progress', () => {
    sessionStorage.setItem(RETURN_URL_KEY, '/progress')
    expect(consumeReturnUrl()).toBe('/progress')
    expect(sessionStorage.getItem(RETURN_URL_KEY)).toBeNull()
  })

  it('returns and consumes url for safe path /dashboard', () => {
    sessionStorage.setItem(RETURN_URL_KEY, '/dashboard')
    expect(consumeReturnUrl()).toBe('/dashboard')
  })

  it('returns url with query for /lessons/123', () => {
    sessionStorage.setItem(RETURN_URL_KEY, '/lessons/abc?foo=1')
    expect(consumeReturnUrl()).toBe('/lessons/abc?foo=1')
  })

  it('returns url for /modules/xyz', () => {
    sessionStorage.setItem(RETURN_URL_KEY, '/modules/xyz')
    expect(consumeReturnUrl()).toBe('/modules/xyz')
  })

  it('returns null for unsafe path and still removes from storage', () => {
    sessionStorage.setItem(RETURN_URL_KEY, '/admin')
    expect(consumeReturnUrl()).toBeNull()
    expect(sessionStorage.getItem(RETURN_URL_KEY)).toBeNull()
  })

  it('returns null for path that only looks like safe (e.g. /lessonsx)', () => {
    sessionStorage.setItem(RETURN_URL_KEY, '/lessonsx')
    expect(consumeReturnUrl()).toBeNull()
  })
})

function mockLocalStorage() {
  const store: Record<string, string> = {}
  const orig = window.localStorage
  Object.defineProperty(window, 'localStorage', {
    value: {
      getItem: (key: string) => store[key] ?? null,
      setItem: (key: string, value: string) => { store[key] = value },
      removeItem: (key: string) => { delete store[key] },
    },
    writable: true,
  })
  return () => {
    Object.defineProperty(window, 'localStorage', { value: orig, writable: true })
  }
}

describe('api', () => {
  let restoreLocalStorage: () => void

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
    restoreLocalStorage = mockLocalStorage()
  })

  afterEach(() => {
    restoreLocalStorage()
    vi.unstubAllGlobals()
  })

  it('throws message on 429', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(null, { status: 429 })
    )
    await expect(api('/test')).rejects.toThrow(
      'Слишком много запросов. Подожди немного и попробуй снова.'
    )
  })

  it('throws server error message from body on 400', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ error: 'Неверный логин' }), { status: 400 })
    )
    await expect(api('/test')).rejects.toThrow('Неверный логин')
  })

  it('throws generic message when body has no error/message', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({}), { status: 500 })
    )
    await expect(api('/test')).rejects.toThrow('Ошибка сервера. Попробуйте позже.')
  })

  it('throws on 401 and clears token', async () => {
    const removeItem = vi.fn()
    const store: Record<string, string> = { token: 'old-token' }
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: (key: string) => store[key] ?? null,
        setItem: vi.fn(),
        removeItem,
      },
      writable: true,
    })
    const origLocation = window.location
    const location = { href: '', pathname: '/progress', search: '' }
    Object.defineProperty(window, 'location', { value: location, writable: true })

    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 401 }))

    await expect(api('/test')).rejects.toThrow('Сессия истекла')

    expect(removeItem).toHaveBeenCalledWith('token')
    expect(location.href).toBe('/login')

    Object.defineProperty(window, 'location', { value: origLocation, writable: true })
  })

  it('returns data on 200', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ id: '1', name: 'Test' }), { status: 200 })
    )
    const result = await api<{ id: string; name: string }>('/test')
    expect(result).toEqual({ id: '1', name: 'Test' })
  })
})

describe('apiVoid', () => {
  let restoreLocalStorage: () => void

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
    restoreLocalStorage = mockLocalStorage()
  })

  afterEach(() => {
    restoreLocalStorage()
    vi.unstubAllGlobals()
  })

  it('does not expect body on 204', async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 204 }))
    await expect(apiVoid('/test', { method: 'DELETE' })).resolves.toBeUndefined()
  })
})
