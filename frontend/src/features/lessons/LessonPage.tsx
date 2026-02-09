import { useCallback, useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { lessons as lessonsApi, progress as progressApi, modules as modulesApi } from '../../shared/api'
import { useAuth } from '../auth/AuthContext'
import { LedBlinkSimulator } from './LedBlinkSimulator'
import { LessonChat } from './LessonChat'
import { MermaidDiagram } from './MermaidDiagram'
import type { Lesson as LessonType, LessonMaterial, UserProgress, Module } from '../../shared/types'

export function LessonPage() {
  const { id } = useParams<{ id: string }>()
  const [lesson, setLesson] = useState<LessonType | null>(null)
  const [module_, setModule] = useState<Module | null>(null)
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const { user } = useAuth()

  const load = useCallback(() => {
    if (!id) return
    setLoading(true)
    Promise.all([
      lessonsApi.get(id),
      user ? progressApi.get().catch(() => null) : Promise.resolve(null),
    ])
      .then(([l, p]) => {
        setLesson(l)
        setProgress(p || null)
        if (l?.module_id) {
          modulesApi.get(l.module_id).then(setModule).catch(() => setModule(null))
        }
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [id, user])

  useEffect(() => {
    load()
  }, [load])

  const setLessonStatus = useCallback(
    (status: string) => {
      if (!id || !user) return
      progressApi.setLesson(id, status).then(() => load())
    },
    [id, user, load]
  )

  const toggleChecklist = useCallback(
    (itemId: string, completed: boolean) => {
      if (!id || !user) return
      progressApi.setChecklist(id, itemId, completed).then(() => load())
    },
    [id, user, load]
  )

  const getLessonStatus = () => {
    if (!progress || !id) return 'not_started'
    const item = progress?.lessons?.find((l) => l.lesson_id === id)
    return item?.status ?? 'not_started'
  }

  const isChecklistDone = (itemId: string) => {
    return progress?.checklist?.some((c) => c.checklist_item_id === itemId && c.completed_at) ?? false
  }

  if (loading || !lesson) {
    if (error) return <p className="error">{error}</p>
    return <p>Загрузка...</p>
  }

  const status = getLessonStatus()
  const orderedLessons = (module_?.lessons ?? []).sort((a, b) => a.sort_order - b.sort_order)
  const currentIndex = orderedLessons.findIndex((l) => l.id === lesson.id)
  const prevLesson = currentIndex > 0 ? orderedLessons[currentIndex - 1] : null
  const nextLesson = currentIndex >= 0 && currentIndex < orderedLessons.length - 1 ? orderedLessons[currentIndex + 1] : null

  const CourseNav = () => (
    <nav className="course-nav" aria-label="Навигация по курсу">
      <Link to={module_ ? `/modules/${module_.id}` : '/'} className="course-nav-home">
        Домой
      </Link>
      {prevLesson ? (
        <Link to={`/lessons/${prevLesson.id}`} className="course-nav-prev">
          ← Назад
        </Link>
      ) : (
        <span className="course-nav-prev disabled" aria-hidden>← Назад</span>
      )}
      {nextLesson ? (
        <Link to={`/lessons/${nextLesson.id}`} className="course-nav-next">
          Вперёд →
        </Link>
      ) : (
        <span className="course-nav-next disabled" aria-hidden>Вперёд →</span>
      )}
    </nav>
  )

  const materials = lesson.materials ?? []
  const sortedMaterials = [...materials].sort((a, b) => {
    const order = (k: string) => (k === 'simulator' ? 0 : k === 'mermaid' ? 1 : 2)
    return order(a.kind) - order(b.kind)
  })
  const simulatorMaterials = sortedMaterials.filter((m) => m.kind === 'simulator' && m.url_or_path === 'led-blink')
  const otherMaterials = sortedMaterials.filter((m) => !(m.kind === 'simulator' && m.url_or_path === 'led-blink'))

  const renderMaterial = (m: LessonMaterial) => {
    if (m.kind === 'simulator' && m.url_or_path === 'led-blink') {
      return (
        <li key={m.id} className="material-simulator">
          <div className="simulator-wrapper" style={{ width: '100%' }}>
            <LedBlinkSimulator />
          </div>
        </li>
      )
    }
    if (m.kind === 'mermaid') {
      return (
        <li key={m.id} className="material-mermaid">
          <MermaidDiagram code={m.url_or_path} title={m.title} />
        </li>
      )
    }
    const isImage = /\.(jpg|jpeg|png|gif|webp|svg)(\?|$)/i.test(m.url_or_path)
    if (m.kind === 'link' && isImage) {
      return (
        <li key={m.id}>
          <figure className="material-image">
            <a href={m.url_or_path} target="_blank" rel="noreferrer">
              <img src={m.url_or_path} alt={m.title || 'Иллюстрация'} loading="lazy" />
            </a>
            {m.title && <figcaption>{m.title}</figcaption>}
          </figure>
        </li>
      )
    }
    return (
      <li key={m.id}>
        {m.kind === 'link' ? (
          <a href={m.url_or_path} target="_blank" rel="noreferrer">
            {m.title || m.url_or_path}
          </a>
        ) : (
          <span>{m.title || m.url_or_path}</span>
        )}
      </li>
    )
  }

  return (
    <div className="lesson-page">
      <CourseNav />
      <div className="lesson-page-content">
      <h1>{lesson.title}</h1>
      {lesson.description && <p className="description">{lesson.description}</p>}
      <p className="meta">Тип: {lesson.lesson_type}</p>
      {lesson.tags && lesson.tags.length > 0 && (
        <p className="tags">Теги: {lesson.tags.join(', ')}</p>
      )}

      {user && (
        <div className="progress-actions">
          <span>Статус: </span>
          <button
            className={status === 'not_started' ? 'active' : ''}
            onClick={() => setLessonStatus('not_started')}
          >
            Не начат
          </button>
          <button
            className={status === 'in_progress' ? 'active' : ''}
            onClick={() => setLessonStatus('in_progress')}
          >
            В процессе
          </button>
          <button
            className={status === 'completed' ? 'active' : ''}
            onClick={() => setLessonStatus('completed')}
          >
            Выполнен
          </button>
        </div>
      )}

      {lesson.steps && lesson.steps.length > 0 && (
        <section className="steps">
          <h2>Шаги</h2>
          {lesson.steps
            .sort((a, b) => a.sort_order - b.sort_order)
            .map((step) => (
              <div key={step.id} className="step">
                <h3>{step.title}</h3>
                <div className="step-content">{step.content}</div>
              </div>
            ))}
        </section>
      )}

      {lesson.checklist && lesson.checklist.length > 0 && (
        <section className="checklist">
          <h2>Чек-лист</h2>
          <ul>
            {lesson.checklist
              .sort((a, b) => a.sort_order - b.sort_order)
              .map((item) => (
                <li key={item.id}>
                  <label>
                    <input
                      type="checkbox"
                      checked={isChecklistDone(item.id)}
                      onChange={(e) => toggleChecklist(item.id, e.target.checked)}
                      disabled={!user}
                    />
                    {item.title}
                  </label>
                </li>
              ))}
          </ul>
        </section>
      )}

      {materials.length > 0 && (
        <section className="materials">
          <h2>Материалы</h2>
          <div className="materials-inner">
            {simulatorMaterials.length > 0 && (
              <ul className="materials-list materials-list-trainer" aria-label="Тренажёр">
                {simulatorMaterials.map(renderMaterial)}
              </ul>
            )}
            {otherMaterials.length > 0 && (
              <ul className="materials-list materials-list-rest">
                {otherMaterials.map(renderMaterial)}
              </ul>
            )}
          </div>
        </section>
      )}

      <LessonChat lesson={lesson} />
      </div>
      <CourseNav />
    </div>
  )
}
