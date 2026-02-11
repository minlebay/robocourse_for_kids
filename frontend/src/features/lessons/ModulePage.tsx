import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { lessons as lessonsApi, modules } from '../../shared/api'
import { useAuth } from '../auth/AuthContext'
import type { Module } from '../../shared/types'
import { MarkdownStepEditor } from './MarkdownStepEditor'
import { PageWithToc, type TocItem } from './PageWithToc'

export function ModulePage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [module_, setModule] = useState<Module | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deletingLessonId, setDeletingLessonId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [createTitle, setCreateTitle] = useState('')
  const [createDesc, setCreateDesc] = useState('')
  const [createType, setCreateType] = useState<'theory' | 'practice' | 'project'>('theory')
  const [createSteps, setCreateSteps] = useState<{ title: string; content: string }[]>([{ title: '', content: '' }])
  const [createError, setCreateError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = useCallback(() => {
    if (!id) return
    modules
      .get(id)
      .then(setModule)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => {
    load()
  }, [load])

  const handleCreate = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      if (!id || !createTitle.trim() || submitting) return
      const steps = createSteps.filter((s) => s.title.trim()).map((s) => ({ title: s.title.trim(), content: s.content.trim() || undefined }))
      setCreateError('')
      setSubmitting(true)
      modules
        .createLesson(id, {
          title: createTitle.trim(),
          description: createDesc.trim() || undefined,
          lesson_type: createType,
          steps: steps.length > 0 ? steps : undefined,
        })
        .then((lesson) => {
          setModule((prev) =>
            prev ? { ...prev, lessons: [...(prev.lessons ?? []), lesson] } : prev
          )
          setCreateTitle('')
          setCreateDesc('')
          setCreateSteps([{ title: '', content: '' }])
          setShowCreateForm(false)
        })
        .catch((err) => setCreateError(err.message))
        .finally(() => setSubmitting(false))
    },
    [id, createTitle, createDesc, createType, createSteps, submitting]
  )

  const addStep = useCallback(() => {
    setCreateSteps((prev) => [...prev, { title: '', content: '' }])
  }, [])

  const updateStep = useCallback((idx: number, field: 'title' | 'content', value: string) => {
    setCreateSteps((prev) => {
      const next = [...prev]
      next[idx] = { ...next[idx], [field]: value }
      return next
    })
  }, [])

  const removeStep = useCallback((idx: number) => {
    setCreateSteps((prev) => (prev.length > 1 ? prev.filter((_, i) => i !== idx) : prev))
  }, [])

  const handleDelete = useCallback(() => {
    if (!id || !user || deleting) return
    const lessonCount = (module_?.lessons ?? []).length
    const msg =
      lessonCount > 0
        ? `Удалить курс «${module_?.title}» и все ${lessonCount} уроков?`
        : `Удалить курс «${module_?.title}»?`
    if (!window.confirm(msg)) return
    setDeleting(true)
    modules
      .delete(id)
      .then(() => navigate('/'))
      .catch((err) => {
        setError(err.message)
        setDeleting(false)
      })
  }, [id, user, deleting, module_, navigate])

  const handleDeleteLesson = useCallback(
    (lessonId: string, lessonTitle: string) => {
      if (!user || deletingLessonId) return
      if (!window.confirm(`Удалить урок «${lessonTitle}»?`)) return
      setDeletingLessonId(lessonId)
      lessonsApi
        .delete(lessonId)
        .then(() => {
          setModule((prev) =>
            prev ? { ...prev, lessons: (prev.lessons ?? []).filter((l) => l.id !== lessonId) } : prev
          )
        })
        .catch((err) => setError(err.message))
        .finally(() => setDeletingLessonId(null))
    },
    [user, deletingLessonId]
  )

  if (loading || !module_) {
    if (error) return <p className="error">{error}</p>
    return <p>Загрузка...</p>
  }

  const orderedLessons = (module_.lessons ?? []).sort((a, b) => a.sort_order - b.sort_order)
  const tocItems: TocItem[] = orderedLessons.map((lesson) => ({
    id: lesson.id,
    title: lesson.title,
    href: `/lessons/${lesson.id}`,
  }))

  return (
    <PageWithToc title={module_.title} items={tocItems}>
      <div className="module-page">
        <p><Link to="/">← Каталог</Link></p>
        <h1>{module_.title}</h1>
        {module_.description && <p>{module_.description}</p>}
        {user?.role === 'teacher' && (
          <div className="module-teacher-actions">
            {!showCreateForm ? (
              <div className="module-teacher-buttons">
                <button type="button" className="button-primary" onClick={() => setShowCreateForm(true)}>
                  Добавить урок
                </button>
                <button
                  type="button"
                  className="button-danger-outline"
                  onClick={handleDelete}
                  disabled={deleting}
                  title="Удалить курс"
                >
                  {deleting ? '…' : 'Удалить курс'}
                </button>
              </div>
            ) : (
              <form className="module-create-lesson-form" onSubmit={handleCreate}>
                <h3>Новый урок</h3>
                {createError && <p className="error">{createError}</p>}
                <div className="form-group">
                  <label htmlFor="lesson-title">Название</label>
                  <input
                    id="lesson-title"
                    value={createTitle}
                    onChange={(e) => setCreateTitle(e.target.value)}
                    placeholder="Название урока"
                    required
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="lesson-desc">Описание</label>
                  <textarea
                    id="lesson-desc"
                    value={createDesc}
                    onChange={(e) => setCreateDesc(e.target.value)}
                    placeholder="Краткое описание (необязательно)"
                    rows={2}
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="lesson-type">Тип</label>
                  <select
                    id="lesson-type"
                    value={createType}
                    onChange={(e) => setCreateType(e.target.value as 'theory' | 'practice' | 'project')}
                  >
                    <option value="theory">Теория</option>
                    <option value="practice">Практика</option>
                    <option value="project">Проект</option>
                  </select>
                </div>
                <div className="form-group">
                  <label>Шаги урока</label>
                  {createSteps.map((step, idx) => (
                    <div key={idx} className="module-create-step">
                      <input
                        value={step.title}
                        onChange={(e) => updateStep(idx, 'title', e.target.value)}
                        placeholder={`Заголовок шага ${idx + 1}`}
                      />
                      <MarkdownStepEditor
                        value={step.content}
                        onChange={(v) => updateStep(idx, 'content', v)}
                        rows={3}
                        placeholder="Контент (Markdown, необязательно)"
                        label=""
                      />
                      <button
                        type="button"
                        className="button-danger-outline"
                        onClick={() => removeStep(idx)}
                        disabled={createSteps.length <= 1}
                      >
                        Удалить
                      </button>
                    </div>
                  ))}
                  <button type="button" className="button-secondary" onClick={addStep}>
                    Добавить шаг
                  </button>
                </div>
                <div className="form-actions">
                  <button type="submit" className="button-primary" disabled={submitting}>
                    Создать
                  </button>
                  <button
                    type="button"
                    className="button-secondary"
                    onClick={() => {
                      setShowCreateForm(false)
                      setCreateTitle('')
                      setCreateDesc('')
                      setCreateSteps([{ title: '', content: '' }])
                      setCreateError('')
                    }}
                  >
                    Отмена
                  </button>
                </div>
              </form>
            )}
          </div>
        )}
        {orderedLessons.length > 0 ? (
          <ul className="lesson-list">
            {orderedLessons.map((lesson) => (
              <li key={lesson.id} className="lesson-list-item">
                <Link to={`/lessons/${lesson.id}`}>{lesson.title}</Link>
                {user?.role === 'teacher' && (
                  <button
                    type="button"
                    className="lesson-list-delete button-danger-outline"
                    onClick={(e) => {
                      e.preventDefault()
                      handleDeleteLesson(lesson.id, lesson.title)
                    }}
                    disabled={deletingLessonId === lesson.id}
                    title="Удалить урок"
                  >
                    {deletingLessonId === lesson.id ? '…' : 'Удалить'}
                  </button>
                )}
              </li>
            ))}
          </ul>
        ) : (
          <p>В этом модуле пока нет уроков.</p>
        )}
      </div>
    </PageWithToc>
  )
}
