import { useCallback, useEffect, useState } from 'react'
import { lessons as lessonsApi, progress as progressApi, modules as modulesApi } from '../../../shared/api'
import type { Lesson, UserProgress, Module } from '../../../shared/types'

export function useLessonData(lessonId: string | undefined, isAuthenticated: boolean) {
  const [lesson, setLesson] = useState<Lesson | null>(null)
  const [module_, setModule] = useState<Module | null>(null)
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    if (!lessonId) return
    setLoading(true)
    Promise.all([
      lessonsApi.get(lessonId),
      isAuthenticated ? progressApi.get().catch(() => null) : Promise.resolve(null),
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
  }, [lessonId, isAuthenticated])

  useEffect(() => {
    load()
  }, [load])

  return { lesson, module_, progress, loading, error, load, setLesson, setProgress, setError }
}
