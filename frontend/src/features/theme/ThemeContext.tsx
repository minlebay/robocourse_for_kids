import {
  createContext,
  useCallback,
  useContext,
  useLayoutEffect,
  useState,
  type ReactNode,
} from 'react'
import {
  themeOrder,
  themes,
  THEME_STORAGE_KEY,
  type ThemeId,
} from './themes'

interface ThemeContextValue {
  themeId: ThemeId
  setTheme: (id: ThemeId) => void
  cycleTheme: () => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

function applyTheme(themeId: ThemeId) {
  const theme = themes[themeId]
  const root = document.documentElement
  if (!theme) return
  for (const [key, value] of Object.entries(theme.cssVars)) {
    root.style.setProperty(key, value)
  }
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [themeId, setThemeIdState] = useState<ThemeId>(() => {
    const saved = localStorage.getItem(THEME_STORAGE_KEY)
    if (saved && themeOrder.includes(saved as ThemeId)) {
      return saved as ThemeId
    }
    return 'default'
  })

  useLayoutEffect(() => {
    applyTheme(themeId)
  }, [themeId])

  const setTheme = useCallback((id: ThemeId) => {
    setThemeIdState(id)
    localStorage.setItem(THEME_STORAGE_KEY, id)
  }, [])

  const cycleTheme = useCallback(() => {
    const idx = themeOrder.indexOf(themeId)
    const nextIdx = (idx + 1) % themeOrder.length
    const next = themeOrder[nextIdx]
    setTheme(next)
  }, [themeId])

  return (
    <ThemeContext.Provider value={{ themeId, setTheme, cycleTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider')
  return ctx
}
