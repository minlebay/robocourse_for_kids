import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { progress as progressApi } from '../../shared/api'
import type { UserProgress } from '../../shared/types'

export function ProgressPage() {
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    progressApi.get()
      .then(setProgress)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <p>Загрузка...</p>
  if (error) return <p className="error">{error}</p>
  if (!progress) return <p>Нет данных о прогрессе.</p>

  const byStatus = (status: string) =>
    progress.lessons.filter((l) => l.status === status)

  return (
    <div className="progress-page">
      <h1>Мой прогресс</h1>
      <p>Выполнено уроков: {byStatus('completed').length}</p>
      <p>В процессе: {byStatus('in_progress').length}</p>
      <section>
        <h2>По статусам</h2>
        <ul>
          {progress.lessons.map((l) => (
            <li key={l.lesson_id}>
              <Link to={`/lessons/${l.lesson_id}`}>
                Урок {l.lesson_id.slice(0, 8)}…
              </Link>
              — {l.status}
            </li>
          ))}
        </ul>
      </section>
      <p>Отмечено пунктов чек-листа: {progress.checklist?.length ?? 0}</p>
    </div>
  )
}
