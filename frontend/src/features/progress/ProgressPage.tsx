import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { progress as progressApi } from '../../shared/api'
import type { UserProgress, LessonStatus } from '../../shared/types'

export function ProgressPage() {
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const { t } = useTranslation()

  useEffect(() => {
    progressApi.get()
      .then(setProgress)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <p>{t('common.loading')}</p>
  if (error) return <p className="error">{error}</p>
  if (!progress) return <p>{t('progress.noData')}</p>

  const byStatus = (status: LessonStatus) =>
    progress.lessons.filter((l) => l.status === status)
  const statusLabel = (status: LessonStatus) =>
    t(`progress.status_${status}` as 'progress.status_not_started')

  return (
    <div className="progress-page">
      <h1>{t('progress.title')}</h1>
      <p>{t('progress.completedCount', { count: byStatus('completed').length })}</p>
      <p>{t('progress.inProgressCount', { count: byStatus('in_progress').length })}</p>
      <section>
        <h2>{t('progress.byStatus')}</h2>
        <ul>
          {progress.lessons.map((l) => (
            <li key={l.lesson_id}>
              <Link to={`/lessons/${l.lesson_id}`}>
                {l.lesson_title || t('progress.lessonFallback', { id: l.lesson_id.slice(0, 8) })}
              </Link>
              {' — '}{statusLabel(l.status)}
            </li>
          ))}
        </ul>
      </section>
      <p>{t('progress.checklistCount', { count: progress.checklist?.length ?? 0 })}</p>
    </div>
  )
}
