import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { lessons as lessonsApi } from '../../../shared/api'
import type { Lesson } from '../../../shared/types'

export type EditStep = { id: string; title: string; content: string }

export type EditForm = {
  title: string
  description: string
  steps: EditStep[]
}

type UseLessonEditArgs = {
  lessonId: string | undefined
  lesson: Lesson | null
  setLesson: React.Dispatch<React.SetStateAction<Lesson | null>>
}

export function useLessonEdit({ lessonId, lesson, setLesson }: UseLessonEditArgs) {
  const { t } = useTranslation()
  const [editing, setEditing] = useState(false)
  const [editForm, setEditForm] = useState<EditForm | null>(null)
  const [saveError, setSaveError] = useState('')

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

  const saveLesson = useCallback(() => {
    if (!lessonId || !editForm) return
    if (!editForm.title.trim()) {
      setSaveError(t('validation.lesson_title_required'))
      return
    }
    if (editForm.steps.some((s) => !s.title.trim())) {
      setSaveError(t('validation.step_title_required'))
      return
    }
    setSaveError('')
    lessonsApi
      .update(lessonId, {
        title: editForm.title,
        description: editForm.description || undefined,
        steps: editForm.steps.map((s, i) => ({ title: s.title, content: s.content, sort_order: i })),
      })
      .then((updated) => {
        setLesson(updated)
        setEditing(false)
        setEditForm(null)
      })
      .catch((err) => setSaveError(err.message))
  }, [lessonId, editForm, t, setLesson])

  const addStep = useCallback(() => {
    setEditForm((prev) =>
      prev ? { ...prev, steps: [...prev.steps, { id: `new-${Date.now()}`, title: '', content: '' }] } : prev
    )
  }, [])

  const removeStep = useCallback((stepId: string) => {
    setEditForm((prev) =>
      prev && prev.steps.length > 1 ? { ...prev, steps: prev.steps.filter((s) => s.id !== stepId) } : prev
    )
  }, [])

  const updateEditStep = useCallback((stepId: string, field: 'title' | 'content', value: string) => {
    setEditForm((prev) => {
      if (!prev) return prev
      const steps = prev.steps.map((s) => (s.id === stepId ? { ...s, [field]: value } : s))
      return { ...prev, steps }
    })
  }, [])

  const setFormTitle = useCallback((value: string) => {
    setEditForm((prev) => (prev ? { ...prev, title: value } : prev))
  }, [])

  const setFormDescription = useCallback((value: string) => {
    setEditForm((prev) => (prev ? { ...prev, description: value } : prev))
  }, [])

  return {
    editing,
    editForm,
    saveError,
    setSaveError,
    startEditing,
    cancelEditing,
    saveLesson,
    addStep,
    removeStep,
    updateEditStep,
    setFormTitle,
    setFormDescription,
  }
}
