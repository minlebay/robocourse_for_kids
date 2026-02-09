import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { modules } from '../../shared/api'
import type { Module } from '../../shared/types'

export function ModulePage() {
  const { id } = useParams<{ id: string }>()
  const [module_, setModule] = useState<Module | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    modules
      .get(id)
      .then(setModule)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [id])

  if (loading || !module_) {
    if (error) return <p className="error">{error}</p>
    return <p>Загрузка...</p>
  }

  return (
    <div className="module-page">
      <p><Link to="/">← Каталог</Link></p>
      <h1>{module_.title}</h1>
      {module_.description && <p>{module_.description}</p>}
      {module_.lessons && module_.lessons.length > 0 ? (
        <ul className="lesson-list">
          {module_.lessons
            .sort((a, b) => a.sort_order - b.sort_order)
            .map((lesson) => (
              <li key={lesson.id}>
                <Link to={`/lessons/${lesson.id}`}>{lesson.title}</Link>
              </li>
            ))}
        </ul>
      ) : (
        <p>В этом модуле пока нет уроков.</p>
      )}
    </div>
  )
}
