import { useCallback, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'

const STORAGE_KEY = 'learn-kids-toc-width'
const MIN_WIDTH = 200
const MAX_WIDTH = 480
const DEFAULT_WIDTH = 280

export type TocItem = {
  id: string
  title: string
  href: string
  active?: boolean
}

type PageWithTocProps = {
  title: string
  items: TocItem[]
  children: React.ReactNode
}

export function PageWithToc({ title, items, children }: PageWithTocProps) {
  const [width, setWidth] = useState(DEFAULT_WIDTH)
  const [isDragging, setIsDragging] = useState(false)
  const sidebarRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY)
      if (saved) {
        const n = parseInt(saved, 10)
        if (n >= MIN_WIDTH && n <= MAX_WIDTH) setWidth(n)
      }
    } catch {
      /* ignore */
    }
  }, [])

  const saveWidth = useCallback((w: number) => {
    const value = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, w))
    setWidth(value)
    try {
      localStorage.setItem(STORAGE_KEY, String(value))
    } catch {
      /* ignore */
    }
  }, [])

  const startDrag = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  useEffect(() => {
    if (!isDragging) return
    const onMove = (e: MouseEvent) => {
      saveWidth(e.clientX)
    }
    const onUp = () => setIsDragging(false)
    document.addEventListener('mousemove', onMove)
    document.addEventListener('mouseup', onUp)
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
    return () => {
      document.removeEventListener('mousemove', onMove)
      document.removeEventListener('mouseup', onUp)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
  }, [isDragging, saveWidth])

  return (
    <div className="content-with-toc">
      <aside
        ref={sidebarRef}
        className="toc-sidebar"
        style={{ width: `${width}px` }}
        aria-label="Содержание"
      >
        <div className="toc-sidebar-inner">
          <h2 className="toc-title">{title}</h2>
          <nav className="toc-nav">
            <ul className="toc-list">
              {items.map((item) => (
                <li key={item.id} className="toc-item">
                  {item.active ? (
                    <span className="toc-link toc-link-active" aria-current="page">
                      {item.title}
                    </span>
                  ) : (
                    <Link to={item.href} className="toc-link">
                      {item.title}
                    </Link>
                  )}
                </li>
              ))}
            </ul>
          </nav>
        </div>
      </aside>
      <div
        className="toc-resize-handle"
        onMouseDown={startDrag}
        role="separator"
        aria-orientation="vertical"
        aria-valuenow={width}
        aria-valuemin={MIN_WIDTH}
        aria-valuemax={MAX_WIDTH}
        title="Изменить ширину содержания"
      />
      <div className="toc-main">{children}</div>
    </div>
  )
}
