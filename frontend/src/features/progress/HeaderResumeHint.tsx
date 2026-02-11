import { useCallback, useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { progress as progressApi } from '../../shared/api'
import type { LessonProgressItem, UserProgress } from '../../shared/types'

/**
 * Находит урок, где пользователь остановился: последний in_progress по updated_at.
 */
function findResumeLesson(progress: UserProgress): LessonProgressItem | null {
  const inProgress = progress.lessons.filter((l) => l.status === 'in_progress')
  if (inProgress.length === 0) return null
  return inProgress.reduce((latest, cur) =>
    new Date(cur.updated_at) > new Date(latest.updated_at) ? cur : latest
  )
}

export function HeaderResumeHint() {
  const { pathname } = useLocation()
  const [resumeLesson, setResumeLesson] = useState<LessonProgressItem | null>(null)

  const load = useCallback(() => {
    progressApi
      .get()
      .then((p) => setResumeLesson(findResumeLesson(p)))
      .catch(() => setResumeLesson(null))
  }, [])

  useEffect(() => {
    load()
  }, [load, pathname])

  if (!resumeLesson?.lesson_title) return null

  return (
    <Link to={`/lessons/${resumeLesson.lesson_id}`} className="header-resume-hint">
      Продолжить: {resumeLesson.lesson_title}
    </Link>
  )
}
