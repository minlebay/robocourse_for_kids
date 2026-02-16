import { useCallback, useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'
import remarkGfm from 'remark-gfm'
import { lessons as lessonsApi, progress as progressApi, modules as modulesApi } from '../../shared/api'
import { useAuth } from '../auth/AuthContext'
import { LedBlinkSimulator } from './LedBlinkSimulator'
import { LessonChat } from './LessonChat'
import { LessonComments } from './LessonComments'
import { MermaidDiagram } from './MermaidDiagram'
import { MarkdownStepEditor } from './MarkdownStepEditor'
import { markdownComponents } from './YouTubeEmbed'
import { PageWithToc, type TocItem } from './PageWithToc'
import type { Lesson as LessonType, LessonMaterial, LessonStatus, UserProgress, Module } from '../../shared/types'
import { getSafeUrl } from '../../shared/url'
import { ConfirmModal } from '../../components'
import { CourseNav } from './CourseNav'
import { LessonProgressBar } from './LessonProgressBar'

type EditStep = { id: string; title: string; content: string }

type EditForm = {
  title: string
  description: string
  steps: EditStep[]
}

export function LessonPage() {
  const { id } = useParams<{ id: string }>()
  const [lesson, setLesson] = useState<LessonType | null>(null)
  const [module_, setModule] = useState<Module | null>(null)
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState(false)
  const [editForm, setEditForm] = useState<EditForm | null>(null)
  const [saveError, setSaveError] = useState('')
  const [deleting, setDeleting] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const { user } = useAuth()
  const navigate = useNavigate()

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
    (status: LessonStatus) => {
      if (!id || !user) return
      progressApi.setLesson(id, status).then(() => load()).catch((err) => setError(err.message))
    },
    [id, user, load]
  )

  const toggleChecklist = useCallback(
    (itemId: string, completed: boolean) => {
      if (!id || !user) return
      if (!progress) {
        progressApi.setChecklist(id, itemId, completed).then(() => load())
        return
      }
      // Оптимистичное обновление — без перезагрузки страницы
      setProgress((prev) => {
        if (!prev) return prev
        const checklist = prev.checklist ?? []
        const existing = checklist.findIndex((c) => c.checklist_item_id === itemId)
        let next: typeof checklist
        if (existing >= 0) {
          next = [...checklist]
          if (completed) {
            next[existing] = { ...next[existing], completed_at: new Date().toISOString() }
          } else {
            next[existing] = { ...next[existing], completed_at: undefined }
          }
        } else if (completed) {
          next = [...checklist, { checklist_item_id: itemId, completed_at: new Date().toISOString() }]
        } else {
          next = checklist.filter((c) => c.checklist_item_id !== itemId)
        }
        return { ...prev, checklist: next }
      })
      progressApi.setChecklist(id, itemId, completed).catch((err) => {
        setError(err.message)
        load() // при ошибке — перезагружаем для восстановления состояния
      })
    },
    [id, user, progress, load]
  )

  const getLessonStatus = (): LessonStatus => {
    if (!progress || !id) return 'not_started'
    const item = progress.lessons?.find((l) => l.lesson_id === id)
    return item?.status ?? 'not_started'
  }

  const isChecklistDone = (itemId: string) => {
    return progress?.checklist?.some((c) => c.checklist_item_id === itemId && c.completed_at) ?? false
  }

  const startEditing = useCallback(() => {
    if (!lesson) return
    const steps = [...(lesson.steps ?? [])].sort((a, b) => a.sort_order - b.sort_order)
    setEditForm({
      title: lesson.title,
      description: lesson.description ?? '',
      steps: steps.map((s) => ({ id: s.id, title: s.title, content: s.content ?? '' })),
    })
    setEditing(true)
    setSaveError('')
  }, [lesson])

  const cancelEditing = useCallback(() => {
    setEditing(false)
    setEditForm(null)
    setSaveError('')
  }, [])

  const requestDeleteLesson = useCallback(() => {
    if (!id || !lesson || !user || deleting) return
    setConfirmDelete(true)
  }, [id, lesson, user, deleting])

  const doDeleteLesson = useCallback(() => {
    if (!id || !user || deleting) return
    setConfirmDelete(false)
    setDeleting(true)
    lessonsApi
      .delete(id)
      .then(() => navigate(module_ ? `/modules/${module_.id}` : '/'))
      .catch((err) => {
        setSaveError(err.message)
        setDeleting(false)
      })
  }, [id, user, deleting, module_, navigate])

  const saveLesson = useCallback(() => {
    if (!id || !editForm) return
    if (!editForm.title.trim()) {
      setSaveError('Введите название урока')
      return
    }
    if (editForm.steps.some((s) => !s.title.trim())) {
      setSaveError('У каждого шага должен быть заголовок')
      return
    }
    setSaveError('')
    lessonsApi
      .update(id, {
        title: editForm.title,
        description: editForm.description || undefined,
        steps: editForm.steps.map((s, i) => ({ title: s.title, content: s.content, sort_order: i })), // порядок по индексу в форме
      })
      .then((updated) => {
        setLesson(updated)
        setEditing(false)
        setEditForm(null)
      })
      .catch((err) => setSaveError(err.message))
  }, [id, editForm])

  const addStep = useCallback(() => {
    setEditForm((prev) =>
      prev
        ? { ...prev, steps: [...prev.steps, { id: `new-${Date.now()}`, title: 'Новый шаг', content: '' }] }
        : prev
    )
  }, [])

  const removeStep = useCallback((stepId: string) => {
    setEditForm((prev) =>
      prev && prev.steps.length > 1
        ? { ...prev, steps: prev.steps.filter((s) => s.id !== stepId) }
        : prev
    )
  }, [])

  const updateEditStep = useCallback((stepId: string, field: 'title' | 'content', value: string) => {
    setEditForm((prev) => {
      if (!prev) return prev
      const steps = prev.steps.map((s) =>
        s.id === stepId ? { ...s, [field]: value } : s
      )
      return { ...prev, steps }
    })
  }, [])

  if (loading) return <p>Загрузка...</p>
  if (error) return <p className="error">{error}</p>
  if (!lesson) return <p className="error">Урок не найден</p>

  const status = getLessonStatus()
  const orderedLessons = (module_?.lessons ?? []).sort((a, b) => a.sort_order - b.sort_order)
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
    const safeUrl = m.kind === 'link' ? getSafeUrl(m.url_or_path) : null
    const isImage = safeUrl && /\.(jpg|jpeg|png|gif|webp|svg)(\?|$)/i.test(m.url_or_path)
    if (m.kind === 'link' && isImage && safeUrl) {
      return (
        <li key={m.id}>
          <figure className="material-image">
            <a href={safeUrl} target="_blank" rel="noreferrer">
              <img src={safeUrl} alt={m.title || 'Иллюстрация'} loading="lazy" />
            </a>
            {m.title && <figcaption>{m.title}</figcaption>}
          </figure>
        </li>
      )
    }
    return (
      <li key={m.id}>
        {m.kind === 'link' && safeUrl ? (
          <a href={safeUrl} target="_blank" rel="noreferrer">
            {m.title || m.url_or_path}
          </a>
        ) : (
          <span>{m.title || m.url_or_path}</span>
        )}
      </li>
    )
  }

  const tocItems: TocItem[] = orderedLessons.map((l) => ({
    id: l.id,
    title: l.title,
    href: `/lessons/${l.id}`,
    active: l.id === lesson.id,
  }))
  const tocTitle = module_ ? module_.title : 'Урок'

  const lessonContent = (
    <div className={`lesson-page${editing && editForm ? ' lesson-page--editing' : ''}`}>
      <CourseNav module={module_} currentLesson={lesson} />
      <div className="lesson-page-content">
      {user?.role === 'teacher' && !editing && (
        <div className="lesson-edit-bar">
          <button type="button" className="button-primary" onClick={startEditing}>
            Редактировать
          </button>
          <button
            type="button"
            className="button-danger-outline"
            onClick={requestDeleteLesson}
            disabled={deleting}
            title="Удалить урок"
          >
            {deleting ? '…' : 'Удалить урок'}
          </button>
        </div>
      )}

      {editing && editForm ? (
        <section className="lesson-edit-form">
          <h2>Редактирование урока</h2>
          {saveError && <p className="error">{saveError}</p>}
          <div className="form-group">
            <label htmlFor="edit-lesson-title">Название</label>
            <input
              id="edit-lesson-title"
              value={editForm.title}
              onChange={(e) => setEditForm((prev) => prev && { ...prev, title: e.target.value })}
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-lesson-desc">Описание</label>
            <textarea
              id="edit-lesson-desc"
              rows={2}
              value={editForm.description}
              onChange={(e) => setEditForm((prev) => prev && { ...prev, description: e.target.value })}
            />
          </div>
          <h3>Шаги</h3>
          {editForm.steps.map((step, idx) => (
            <div key={step.id} className="lesson-edit-step">
              <div className="form-group">
                <label>Заголовок шага {idx + 1}</label>
                <input
                  value={step.title}
                  onChange={(e) => updateEditStep(step.id, 'title', e.target.value)}
                />
              </div>
              <div className="form-group">
                <MarkdownStepEditor
                  value={step.content}
                  onChange={(v) => updateEditStep(step.id, 'content', v)}
                  label="Контент (Markdown)"
                />
              </div>
              <button
                type="button"
                className="button-danger-outline"
                onClick={() => removeStep(step.id)}
                disabled={editForm.steps.length <= 1}
              >
                Удалить шаг
              </button>
            </div>
          ))}
          <div className="form-actions">
            <button type="button" className="button-secondary" onClick={addStep}>
              Добавить шаг
            </button>
            <button type="button" className="button-primary" onClick={saveLesson}>
              Сохранить
            </button>
            <button type="button" className="button-secondary" onClick={cancelEditing}>
              Отмена
            </button>
          </div>
        </section>
      ) : (
        <>
      <h1>{lesson.title}</h1>
      {lesson.description && <p className="description">{lesson.description}</p>}
      <p className="meta">Тип: {lesson.lesson_type}</p>
      {lesson.tags && lesson.tags.length > 0 && (
        <p className="tags">Теги: {lesson.tags.join(', ')}</p>
      )}

      <LessonProgressBar
        status={status}
        onStatusChange={setLessonStatus}
        isAuthenticated={!!user}
      />

      {lesson.steps && lesson.steps.length > 0 && (
        <section className="steps">
          <h2>Шаги</h2>
          {[...lesson.steps]
            .sort((a, b) => a.sort_order - b.sort_order)
            .map((step) => (
              <div key={step.id} className="step">
                <h3>{step.title}</h3>
                <div className="step-content step-content-markdown">
                  <ReactMarkdown rehypePlugins={[rehypeHighlight]} remarkPlugins={[remarkGfm]} components={markdownComponents}>{step.content ?? ''}</ReactMarkdown>
                </div>
              </div>
            ))}
        </section>
      )}

      {lesson.checklist && lesson.checklist.length > 0 && (
        <section className="checklist">
          <h2>Чек-лист</h2>
          <ul>
            {[...lesson.checklist]
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

      <LessonChat lessonId={lesson.id} />
      <LessonComments lessonId={lesson.id} />
        </>
      )}
      </div>
      <CourseNav module={module_} currentLesson={lesson} />
    </div>
  )

  return (
    <PageWithToc
      title={tocTitle}
      items={tocItems}
      containerClassName={editing && editForm ? 'content-with-toc--editing' : undefined}
    >
      <ConfirmModal
        open={confirmDelete}
        title="Удалить урок?"
        confirmLabel="Удалить"
        variant="danger"
        onConfirm={doDeleteLesson}
        onCancel={() => setConfirmDelete(false)}
      >
        {lesson ? `Удалить урок «${lesson.title}»?` : ''}
      </ConfirmModal>
      {lessonContent}
    </PageWithToc>
  )
}
