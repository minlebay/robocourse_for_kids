import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { lessons as lessonsApi, modules } from '../../shared/api'
import { useAuth } from '../auth/AuthContext'
import type { Module } from '../../shared/types'
import { ConfirmModal } from '../../components'
import { MarkdownStepEditor } from './MarkdownStepEditor'
import { PageWithToc, type TocItem } from './PageWithToc'

type ConfirmKind = { kind: 'module' } | { kind: 'lesson'; lessonId: string; lessonTitle: string }

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
  const [confirm, setConfirm] = useState<ConfirmKind | null>(null)

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
    setConfirm({ kind: 'module' })
  }, [id, user, deleting])

  const doDeleteModule = useCallback(() => {
    if (!id || !confirm || confirm.kind !== 'module') return
    setConfirm(null)
    setDeleting(true)
    modules
      .delete(id)
      .then(() => navigate('/'))
      .catch((err) => {
        setError(err.message)
        setDeleting(false)
      })
  }, [id, confirm, navigate])

  const handleDeleteLesson = useCallback((lessonId: string, lessonTitle: string) => {
    if (!user || deletingLessonId) return
    setConfirm({ kind: 'lesson', lessonId, lessonTitle })
  }, [user, deletingLessonId])

  const doDeleteLesson = useCallback(() => {
    if (!confirm || confirm.kind !== 'lesson') return
    const { lessonId } = confirm
    setConfirm(null)
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
  }, [confirm])

  const confirmTitle = confirm?.kind === 'module'
    ? 'Удалить курс?'
    : confirm?.kind === 'lesson'
      ? 'Удалить урок?'
      : ''
  const confirmMessage =
    confirm?.kind === 'module' && module_
      ? (module_.lessons ?? []).length > 0
        ? `Удалить курс «${module_.title}» и все ${(module_.lessons ?? []).length} уроков?`
        : `Удалить курс «${module_.title}»?`
      : confirm?.kind === 'lesson'
        ? `Удалить урок «${confirm.lessonTitle}»?`
        : ''

  if (loading) return <p>Загрузка...</p>
  if (error) return <p className="error">{error}</p>
  if (!module_) return <p className="error">Модуль не найден</p>

  const orderedLessons = (module_.lessons ?? []).sort((a, b) => a.sort_order - b.sort_order)
  const tocItems: TocItem[] = orderedLessons.map((lesson) => ({
    id: lesson.id,
    title: lesson.title,
    href: `/lessons/${lesson.id}`,
  }))

  return (
    <PageWithToc title={module_.title} items={tocItems}>
      <ConfirmModal
        open={!!confirm}
        title={confirmTitle}
        confirmLabel="Удалить"
        variant="danger"
        onConfirm={confirm?.kind === 'module' ? doDeleteModule : doDeleteLesson}
        onCancel={() => setConfirm(null)}
      >
        {confirmMessage}
      </ConfirmModal>
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
