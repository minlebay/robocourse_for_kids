import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { modules } from '../../shared/api'
import type { Module } from '../../shared/types'
import { PageWithToc, type TocItem } from './PageWithToc'

export function MyCoursesPage() {
  const [list, setList] = useState<Module[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const { t } = useTranslation()

  const load = useCallback(() => {
    modules
      .list({ mine: true })
      .then(setList)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  if (loading) return <p>{t('common.loading')}</p>
  if (error) return <p className="error">{error}</p>

  const sortedList = [...list].sort((a, b) => a.sort_order - b.sort_order)
  const tocItems: TocItem[] = sortedList.map((mod) => ({
    id: mod.id,
    title: mod.title,
    href: `/modules/${mod.id}`,
  }))

  return (
    <PageWithToc title={t('myCourses.title')} items={tocItems}>
      <div className="catalog">
        <h1>{t('myCourses.heading')}</h1>
        {sortedList.length === 0 ? (
          <p>{t('myCourses.empty')}</p>
        ) : (
          <div className="module-list">
            {sortedList.map((mod) => (
              <section key={mod.id} className="module-card">
                <h2>
                  <Link to={`/modules/${mod.id}`}>{mod.title}</Link>
                </h2>
                {mod.description && <p>{mod.description}</p>}
              </section>
            ))}
          </div>
        )}
      </div>
    </PageWithToc>
  )
}
