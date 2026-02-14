import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { progress as progressApi } from '../../shared/api'
import type { LessonProgressItem, UserProgress } from '../../shared/types'

const THROTTLE_MS = 30_000 // не чаще раза в 30 секунд

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
  const lastFetchedAt = useRef(0)

  const load = useCallback(() => {
    const now = Date.now()
    if (now - lastFetchedAt.current < THROTTLE_MS) return
    lastFetchedAt.current = now
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
