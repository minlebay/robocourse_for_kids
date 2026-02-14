import { useEffect, useId, useRef, useState } from 'react'
import mermaid from 'mermaid'

// Безопасность: mermaid.render() с securityLevel: 'strict' возвращает SVG без скриптов и опасных атрибутов.
// Допущение: вывод библиотеки mermaid считаем доверенным; при обновлении библиотеки проверять changelog на безопасность.
mermaid.initialize({
  startOnLoad: false,
  theme: 'dark',
  securityLevel: 'strict',
  themeVariables: {
    primaryColor: '#e07c24',
    primaryTextColor: '#e2e8f0',
    primaryBorderColor: '#334155',
    lineColor: '#64748b',
    secondaryColor: '#1e293b',
    tertiaryColor: '#0f172a',
    background: '#1e293b',
    mainBkgColor: '#1e293b',
    secondBkgColor: '#0f172a',
    textColor: '#e2e8f0',
    border1: '#334155',
    border2: '#475569',
    noteBkgColor: '#0f172a',
    noteTextColor: '#e2e8f0',
    noteBorderColor: '#334155',
  },
})

export function MermaidDiagram({ code, title }: { code: string; title?: string }) {
  const id = useId().replace(/:/g, '')
  const containerRef = useRef<HTMLDivElement>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!code.trim() || !containerRef.current) return
    setError(null)
    const el = containerRef.current
    el.innerHTML = ''
    mermaid
      .render(`mermaid-${id}`, code.trim())
      .then(({ svg }) => {
        if (el) el.innerHTML = svg // доверенный вывод mermaid (securityLevel: 'strict')
      })
      .catch((err) => {
        setError(err.message || 'Ошибка отрисовки диаграммы')
      })
  }, [code, id])

  if (error) {
    return (
      <figure className="mermaid-diagram mermaid-diagram-error">
        {title && <figcaption className="mermaid-diagram-caption">{title}</figcaption>}
        <p className="mermaid-error-text">{error}</p>
      </figure>
    )
  }

  return (
    <figure className="mermaid-diagram">
      {title && <figcaption className="mermaid-diagram-caption">{title}</figcaption>}
      <div className="mermaid-diagram-svg" ref={containerRef} />
    </figure>
  )
}
