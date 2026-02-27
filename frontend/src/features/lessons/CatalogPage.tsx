import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { modules } from '../../shared/api'
import { useAuth } from '../auth/AuthContext'
import { hasTeacherAccess } from '../../shared/roles'
import type { Module } from '../../shared/types'
import { PageWithToc, type TocItem } from './PageWithToc'

export function CatalogPage() {
  const { user } = useAuth()
  const [list, setList] = useState<Module[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [createTitle, setCreateTitle] = useState('')
  const [createDesc, setCreateDesc] = useState('')
  const [createError, setCreateError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const { t } = useTranslation()

  const load = useCallback(() => {
    modules
      .list()
      .then(setList)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const handleCreate = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      if (!createTitle.trim() || submitting) return
      setCreateError('')
      setSubmitting(true)
      modules
        .create({ title: createTitle.trim(), description: createDesc.trim() || undefined })
        .then((mod) => {
          setList((prev) => [...prev, mod])
          setCreateTitle('')
          setCreateDesc('')
          setShowCreateForm(false)
        })
        .catch((err) => setCreateError(err.message))
        .finally(() => setSubmitting(false))
    },
    [createTitle, createDesc, submitting]
  )

  if (loading) return <p>{t('common.loading')}</p>
  if (error) return <p className="error">{error}</p>

  const sortedList = [...list].sort((a, b) => a.sort_order - b.sort_order)
  const tocItems: TocItem[] = sortedList.map((mod) => ({
    id: mod.id,
    title: mod.title,
    href: `/modules/${mod.id}`,
  }))

  return (
    <PageWithToc title={t('catalog.title')} items={tocItems}>
      <div className="catalog">
        <h1>{t('catalog.heading')}</h1>
        {hasTeacherAccess(user) && (
          <div className="catalog-teacher-actions">
            {!showCreateForm ? (
              <button type="button" className="button-primary" onClick={() => setShowCreateForm(true)}>
                {t('catalog.createCourse')}
              </button>
            ) : (
              <form className="catalog-create-form" onSubmit={handleCreate}>
                <h3>{t('catalog.newCourse')}</h3>
                {createError && <p className="error">{createError}</p>}
                <div className="form-group">
                  <label htmlFor="module-title">{t('catalog.courseTitle')}</label>
                  <input
                    id="module-title"
                    value={createTitle}
                    onChange={(e) => setCreateTitle(e.target.value)}
                    placeholder={t('catalog.courseTitlePlaceholder')}
                    required
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="module-desc">{t('catalog.description')}</label>
                  <textarea
                    id="module-desc"
                    value={createDesc}
                    onChange={(e) => setCreateDesc(e.target.value)}
                    placeholder={t('catalog.descriptionPlaceholder')}
                    rows={2}
                  />
                </div>
                <div className="form-actions">
                  <button type="submit" className="button-primary" disabled={submitting}>
                    {t('common.create')}
                  </button>
                  <button
                    type="button"
                    className="button-secondary"
                    onClick={() => {
                      setShowCreateForm(false)
                      setCreateTitle('')
                      setCreateDesc('')
                      setCreateError('')
                    }}
                  >
                    {t('common.cancel')}
                  </button>
                </div>
              </form>
            )}
          </div>
        )}
        <div className="module-list">
          {sortedList.map((mod) => (
            <section key={mod.id} className="module-card">
              <h2><Link to={`/modules/${mod.id}`}>{mod.title}</Link></h2>
              {mod.description && <p>{mod.description}</p>}
            </section>
          ))}
        </div>
      </div>
    </PageWithToc>
  )
}
