import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useState,
  type ReactNode,
} from 'react'
import { auth as authApi } from '../../shared/api'
import { useAuth } from '../auth/AuthContext'
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

function getStoredTheme(): ThemeId | null {
  try {
    const saved = localStorage.getItem(THEME_STORAGE_KEY)
    if (!saved) return null
    const trimmed = saved.trim()
    if (themeOrder.includes(trimmed as ThemeId)) return trimmed as ThemeId
    return null
  } catch {
    return null
  }
}

function setStoredTheme(id: ThemeId): void {
  try {
    localStorage.setItem(THEME_STORAGE_KEY, id)
  } catch {
    // localStorage недоступен (приватный режим, quota и т.д.)
  }
}

function applyTheme(themeId: ThemeId) {
  const theme = themes[themeId]
  const root = document.documentElement
  if (!theme) return
  for (const [key, value] of Object.entries(theme.cssVars)) {
    root.style.setProperty(key, value)
  }
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const { user, setUser } = useAuth()
  const [themeId, setThemeIdState] = useState<ThemeId>(() => {
    return getStoredTheme() ?? 'default'
  })

  // Синхронизация с темой из профиля пользователя при загрузке
  useEffect(() => {
    if (user?.theme && themeOrder.includes(user.theme as ThemeId)) {
      setThemeIdState(user.theme as ThemeId)
      setStoredTheme(user.theme as ThemeId)
    }
  }, [user?.theme])

  useLayoutEffect(() => {
    applyTheme(themeId)
  }, [themeId])

  const setTheme = useCallback(
    async (id: ThemeId) => {
      setThemeIdState(id)
      setStoredTheme(id)
      if (user) {
        try {
          const updated = await authApi.updateTheme(id)
          setUser(updated)
        } catch {
          // Сохранение в профиль не удалось, тема осталась в localStorage
        }
      }
    },
    [user, setUser]
  )

  const cycleTheme = useCallback(() => {
    const idx = themeOrder.indexOf(themeId)
    const nextIdx = (idx + 1) % themeOrder.length
    const next = themeOrder[nextIdx]
    setTheme(next)
  }, [themeId, setTheme])

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
