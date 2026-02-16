import { useTheme } from './ThemeContext'
import { themeOrder, themes } from './themes'

export function ThemeSelector() {
  const { themeId, setTheme } = useTheme()
  const theme = themes[themeId]

  return (
    <div className="theme-selector-wrap" aria-label="Выбор темы">
      <button
        type="button"
        className="theme-selector"
        title={theme.name}
        aria-label={`Тема: ${theme.name}. Наведите для выбора другой темы.`}
        aria-haspopup="listbox"
        aria-expanded="false"
      >
        <span className="theme-selector-preview">
          <span style={{ backgroundColor: theme.previewColors[0] }} />
          <span style={{ backgroundColor: theme.previewColors[1] }} />
          <span style={{ backgroundColor: theme.previewColors[2] }} />
          <span style={{ backgroundColor: theme.previewColors[3] }} />
        </span>
      </button>
      <div className="theme-selector-dropdown" role="listbox" aria-label="Список тем">
        {themeOrder.map((id) => {
          const t = themes[id]
          const active = id === themeId
          return (
            <button
              key={id}
              type="button"
              role="option"
              aria-selected={active}
              className="theme-selector-option"
              onClick={() => setTheme(id)}
            >
              <span className="theme-selector-option-preview">
                <span style={{ backgroundColor: t.previewColors[0] }} />
                <span style={{ backgroundColor: t.previewColors[1] }} />
                <span style={{ backgroundColor: t.previewColors[2] }} />
                <span style={{ backgroundColor: t.previewColors[3] }} />
              </span>
              <span className="theme-selector-option-name">{t.name}</span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
