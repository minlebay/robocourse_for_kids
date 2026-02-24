import { createContext, useContext } from 'react'
import type { ThemeId } from './themes'

export interface ThemeContextValue {
  themeId: ThemeId
  setTheme: (id: ThemeId) => void
  cycleTheme: () => void
}

export const ThemeContext = createContext<ThemeContextValue | null>(null)

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider')
  return ctx
}
