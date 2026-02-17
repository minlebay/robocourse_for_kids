import { useCallback, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import styles from './PageWithToc.module.css'

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
  /** Класс для контейнера (например, при режиме редактирования урока) */
  containerClassName?: string
}

function getInitialTocWidth(): number {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      const n = parseInt(saved, 10)
      if (n >= MIN_WIDTH && n <= MAX_WIDTH) return n
    }
  } catch {
    /* ignore */
  }
  return DEFAULT_WIDTH
}

export function PageWithToc({ title, items, children, containerClassName }: PageWithTocProps) {
  const [width, setWidth] = useState(getInitialTocWidth)
  const [isDragging, setIsDragging] = useState(false)
  const sidebarRef = useRef<HTMLDivElement>(null)
  const dragStartRef = useRef({ clientX: 0, width: 0 })

  const saveWidth = useCallback((w: number) => {
    const value = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, w))
    setWidth(value)
    try {
      localStorage.setItem(STORAGE_KEY, String(value))
    } catch {
      /* ignore */
    }
  }, [])

  const startDrag = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault()
      dragStartRef.current = { clientX: e.clientX, width }
      setIsDragging(true)
    },
    [width]
  )

  useEffect(() => {
    if (!isDragging) return
    const onMove = (e: MouseEvent) => {
      const { clientX: startX, width: startWidth } = dragStartRef.current
      saveWidth(startWidth + (e.clientX - startX))
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

  const isEditing = containerClassName === 'content-with-toc--editing'

  return (
    <div
      className={`${styles.container} ${isEditing ? styles.editing : ''}`}
      role="presentation"
    >
      <aside
        ref={sidebarRef}
        className={styles.sidebar}
        style={{ width: `${width}px` }}
        aria-label="Содержание"
      >
        <div className={styles.sidebarInner}>
          <h2 className={styles.title}>{title}</h2>
          <nav className={styles.nav} aria-label="Навигация по урокам">
            <ul className={styles.list}>
              {items.map((item) => (
                <li key={item.id} className={styles.item}>
                  {item.active ? (
                    <span className={`${styles.link} ${styles.linkActive}`} aria-current="page">
                      {item.title}
                    </span>
                  ) : (
                    <Link to={item.href} className={styles.link}>
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
        className={styles.resizeHandle}
        onMouseDown={startDrag}
        role="separator"
        aria-orientation="vertical"
        aria-valuenow={width}
        aria-valuemin={MIN_WIDTH}
        aria-valuemax={MAX_WIDTH}
        aria-label="Изменить ширину содержания"
        title="Изменить ширину содержания"
      />
      <div className={styles.main}>{children}</div>
    </div>
  )
}
