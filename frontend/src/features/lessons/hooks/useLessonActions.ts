import { useCallback } from 'react'
import { progress as progressApi, lessons as lessonsApi } from '../../../shared/api'
import type { Lesson, LessonStatus, UserProgress, ReactionType } from '../../../shared/types'

type UseLessonActionsArgs = {
  lessonId: string | undefined
  isAuthenticated: boolean
  progress: UserProgress | null
  load: () => void
  setError: (msg: string) => void
  setLesson: React.Dispatch<React.SetStateAction<Lesson | null>>
  setProgress: React.Dispatch<React.SetStateAction<UserProgress | null>>
}

export function useLessonActions({
  lessonId,
  isAuthenticated,
  progress,
  load,
  setError,
  setLesson,
  setProgress,
}: UseLessonActionsArgs) {
  const setLessonStatus = useCallback(
    (status: LessonStatus) => {
      if (!lessonId || !isAuthenticated) return
      progressApi.setLesson(lessonId, status).then(() => load()).catch((err) => setError(err.message))
    },
    [lessonId, isAuthenticated, load, setError]
  )

  const toggleChecklist = useCallback(
    (itemId: string, completed: boolean) => {
      if (!lessonId || !isAuthenticated) return
      if (!progress) {
        progressApi.setChecklist(lessonId, itemId, completed).then(() => load())
        return
      }
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
      progressApi.setChecklist(lessonId, itemId, completed).catch((err) => {
        setError(err.message)
        load()
      })
    },
    [lessonId, isAuthenticated, progress, load, setError, setProgress]
  )

  const setLessonReaction = useCallback(
    (reaction: ReactionType) => {
      if (!lessonId || !isAuthenticated) return
      setLesson((prev) => {
        if (!prev) return prev
        const wasLike = prev.user_reaction === 'like'
        const wasDislike = prev.user_reaction === 'dislike'
        let likes = prev.likes_count ?? 0
        let dislikes = prev.dislikes_count ?? 0
        if (reaction === 'like') {
          if (wasDislike) dislikes -= 1
          if (!wasLike) likes += 1
          return { ...prev, user_reaction: 'like' as const, likes_count: likes, dislikes_count: dislikes }
        }
        if (reaction === 'dislike') {
          if (wasLike) likes -= 1
          if (!wasDislike) dislikes += 1
          return { ...prev, user_reaction: 'dislike' as const, likes_count: likes, dislikes_count: dislikes }
        }
        return prev
      })
      lessonsApi.setReaction(lessonId, reaction).catch((err) => {
        setError(err.message)
        load()
      })
    },
    [lessonId, isAuthenticated, setLesson, load, setError]
  )

  const clearLessonReaction = useCallback(() => {
    if (!lessonId || !isAuthenticated) return
    setLesson((prev) => {
      if (!prev) return prev
      const wasLike = prev.user_reaction === 'like'
      const wasDislike = prev.user_reaction === 'dislike'
      let likes = prev.likes_count ?? 0
      let dislikes = prev.dislikes_count ?? 0
      if (wasLike) likes = Math.max(0, likes - 1)
      if (wasDislike) dislikes = Math.max(0, dislikes - 1)
      return { ...prev, user_reaction: undefined, likes_count: likes, dislikes_count: dislikes }
    })
    lessonsApi.deleteReaction(lessonId).catch((err) => {
      setError(err.message)
      load()
    })
  }, [lessonId, isAuthenticated, setLesson, load, setError])

  const getLessonStatus = useCallback((): LessonStatus => {
    if (!progress || !lessonId) return 'not_started'
    const item = progress.lessons?.find((l) => l.lesson_id === lessonId)
    return item?.status ?? 'not_started'
  }, [progress, lessonId])

  const isChecklistDone = useCallback(
    (itemId: string) => {
      return progress?.checklist?.some((c) => c.checklist_item_id === itemId && c.completed_at) ?? false
    },
    [progress]
  )

  return {
    setLessonStatus,
    toggleChecklist,
    setLessonReaction,
    clearLessonReaction,
    getLessonStatus,
    isChecklistDone,
  }
}
