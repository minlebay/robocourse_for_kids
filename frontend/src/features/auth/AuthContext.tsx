import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import { auth as authApi } from '../../shared/api'
import type { User } from '../../shared/types'

interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (login: string, password: string) => Promise<void>
  register: (login: string, password: string, name: string, role?: string, inviteCode?: string) => Promise<void>
  logout: () => void
  setUser: (u: User | null) => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const loadUser = useCallback(async () => {
    const token = localStorage.getItem('token')
    if (!token) {
      setLoading(false)
      return
    }
    try {
      const u = await authApi.me()
      setUser(u)
    } catch {
      localStorage.removeItem('token')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadUser()
  }, [loadUser])

  const login = useCallback(async (loginName: string, password: string) => {
    const { user: u, token } = await authApi.login({ login: loginName, password })
    localStorage.setItem('token', token)
    setUser(u)
  }, [])

  const register = useCallback(
    async (loginName: string, password: string, name: string, role?: string, inviteCode?: string) => {
      const { user: u, token } = await authApi.register({
        login: loginName,
        password,
        name,
        role,
        invite_code: inviteCode || undefined,
      })
      localStorage.setItem('token', token)
      setUser(u)
    },
    []
  )

  const logout = useCallback(() => {
    localStorage.removeItem('token')
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider
      value={{ user, loading, login, register, logout, setUser }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
