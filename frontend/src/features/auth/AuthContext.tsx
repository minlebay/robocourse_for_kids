import { createContext, useContext } from 'react'
import type { User } from '../../shared/types'

export interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (login: string, password: string) => Promise<void>
  register: (login: string, password: string, name: string, role?: string, inviteCode?: string) => Promise<void>
  logout: () => void
  setUser: (u: User | null) => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
