import { useTheme } from './ThemeContext'
import { themes } from './themes'

export function ThemeSelector() {
  const { themeId, cycleTheme } = useTheme()
  const theme = themes[themeId]

  return (
    <button
      type="button"
      className="theme-selector"
      onClick={cycleTheme}
      title={theme.name}
      aria-label={`Тема: ${theme.name}. Нажмите для переключения.`}
    >
      <span className="theme-selector-preview">
        <span style={{ backgroundColor: theme.previewColors[0] }} />
        <span style={{ backgroundColor: theme.previewColors[1] }} />
        <span style={{ backgroundColor: theme.previewColors[2] }} />
        <span style={{ backgroundColor: theme.previewColors[3] }} />
      </span>
    </button>
  )
}
