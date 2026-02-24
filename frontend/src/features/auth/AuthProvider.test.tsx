import { describe, it, expect, vi, beforeAll, beforeEach, afterAll } from 'vitest'
import { render, screen, waitFor, act } from '@testing-library/react'
import { AuthProvider } from './AuthProvider'
import { useAuth } from './AuthContext'

vi.mock('../../shared/api', () => ({
  auth: {
    me: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
  },
}))

const { auth: authApi } = await import('../../shared/api')
const mockMe = authApi.me as ReturnType<typeof vi.fn>
const mockLogin = authApi.login as ReturnType<typeof vi.fn>

const mockUser = { id: '1', login: 'alice', name: 'Alice', role: 'student' as const, created_at: '' }

// localStorage mock — jsdom в тестовом окружении может не иметь полноценного localStorage
const localStorageMock = (() => {
  const store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { Object.keys(store).forEach((k) => delete store[k]) },
  }
})()

beforeAll(() => vi.stubGlobal('localStorage', localStorageMock))
afterAll(() => vi.unstubAllGlobals())

function TestConsumer() {
  const { user, loading } = useAuth()
  if (loading) return <span>loading</span>
  return <span>{user ? `user:${user.login}` : 'no-user'}</span>
}

beforeEach(() => {
  vi.clearAllMocks()
  localStorageMock.clear()
})

describe('AuthProvider', () => {
  it('sets loading=false and user=null when no token', async () => {
    render(<AuthProvider><TestConsumer /></AuthProvider>)
    await waitFor(() => expect(screen.getByText('no-user')).toBeInTheDocument())
    expect(mockMe).not.toHaveBeenCalled()
  })

  it('calls me() and sets user when token exists', async () => {
    localStorageMock.setItem('token', 'tok')
    mockMe.mockResolvedValue(mockUser)
    render(<AuthProvider><TestConsumer /></AuthProvider>)
    await waitFor(() => expect(screen.getByText('user:alice')).toBeInTheDocument())
  })

  it('removes token and sets user=null when me() fails', async () => {
    localStorageMock.setItem('token', 'bad-tok')
    mockMe.mockRejectedValue(new Error('401'))
    render(<AuthProvider><TestConsumer /></AuthProvider>)
    await waitFor(() => expect(screen.getByText('no-user')).toBeInTheDocument())
    expect(localStorageMock.getItem('token')).toBeNull()
  })

  it('login() stores token and sets user', async () => {
    mockLogin.mockResolvedValue({ user: mockUser, token: 'new-tok' })
    function LoginTrigger() {
      const { login, user } = useAuth()
      return (
        <>
          <button onClick={() => login('alice', 'pass')}>login</button>
          <span>{user ? `user:${user.login}` : 'no-user'}</span>
        </>
      )
    }
    render(<AuthProvider><LoginTrigger /></AuthProvider>)
    await waitFor(() => expect(screen.getByText('no-user')).toBeInTheDocument())
    await act(() => screen.getByRole('button').click())
    expect(localStorageMock.getItem('token')).toBe('new-tok')
    expect(screen.getByText('user:alice')).toBeInTheDocument()
  })

  it('logout() removes token and clears user', async () => {
    localStorageMock.setItem('token', 'tok')
    mockMe.mockResolvedValue(mockUser)
    function LogoutTrigger() {
      const { logout, user } = useAuth()
      return (
        <>
          <button onClick={logout}>logout</button>
          <span>{user ? `user:${user.login}` : 'no-user'}</span>
        </>
      )
    }
    render(<AuthProvider><LogoutTrigger /></AuthProvider>)
    await waitFor(() => expect(screen.getByText('user:alice')).toBeInTheDocument())
    await act(() => screen.getByRole('button').click())
    expect(localStorageMock.getItem('token')).toBeNull()
    expect(screen.getByText('no-user')).toBeInTheDocument()
  })
})
