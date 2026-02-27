import { useCallback, useEffect, useState } from 'react'
import { lessons as lessonsApi, progress as progressApi, modules as modulesApi } from '../../../shared/api'
import type { Lesson, UserProgress, Module } from '../../../shared/types'

export function useLessonData(lessonId: string | undefined, isAuthenticated: boolean) {
  const [lesson, setLesson] = useState<Lesson | null>(null)
  const [module_, setModule] = useState<Module | null>(null)
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [locked, setLocked] = useState(false)
  const [refetchKey, setRefetchKey] = useState(0)

  // Сброс состояния при смене урока во время рендера (паттерн React «adjusting state based on props»).
  // Избегает setState в useEffect для синхронных сбросов.
  const [prevLessonId, setPrevLessonId] = useState(lessonId)
  if (lessonId !== prevLessonId) {
    setPrevLessonId(lessonId)
    setLoading(true)
    setLesson(null)
    setModule(null)
    setProgress(null)
    setError('')
    setLocked(false)
  }

  useEffect(() => {
    if (!lessonId) return
    let cancelled = false

    // Все setState внутри асинхронных колбэков (.then/.catch/.finally) — не синхронно в теле эффекта
    Promise.all([
      lessonsApi.get(lessonId),
      isAuthenticated ? progressApi.get().catch(() => null) : Promise.resolve(null),
    ])
      .then(([l, p]) => {
        if (cancelled) return
        setLesson(l)
        setProgress(p ?? null)
        if (l?.module_id) {
          modulesApi
            .get(l.module_id)
            .then((m) => { if (!cancelled) setModule(m) })
            .catch(() => { if (!cancelled) setModule(null) })
        }
      })
      .catch((err) => {
        if (!cancelled) {
          if (err.message === 'auth_required') {
            setLocked(true)
          } else {
            setError(err.message)
          }
        }
      })
      .finally(() => { if (!cancelled) setLoading(false) })

    return () => { cancelled = true }
  }, [lessonId, isAuthenticated, refetchKey])

  const load = useCallback(() => {
    setLoading(true)
    setError('')
    setRefetchKey((k) => k + 1)
  }, [])

  return { lesson, module_, progress, loading, error, locked, load, setLesson, setProgress, setError }
}
