import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { lessons as lessonsApi } from '../../../shared/api'
import type { Module } from '../../../shared/types'

type UseLessonDeleteArgs = {
  lessonId: string | undefined
  isTeacher: boolean
  module_: Module | null
  setSaveError: (msg: string) => void
}

export function useLessonDelete({ lessonId, isTeacher, module_, setSaveError }: UseLessonDeleteArgs) {
  const navigate = useNavigate()
  const [deleting, setDeleting] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const requestDeleteLesson = useCallback(() => {
    if (!lessonId || !isTeacher || deleting) return
    setConfirmDelete(true)
  }, [lessonId, isTeacher, deleting])

  const doDeleteLesson = useCallback(() => {
    if (!lessonId || !isTeacher || deleting) return
    setConfirmDelete(false)
    setDeleting(true)
    lessonsApi
      .delete(lessonId)
      .then(() => navigate(module_ ? `/modules/${module_.id}` : '/'))
      .catch((err) => {
        setSaveError(err.message)
        setDeleting(false)
      })
  }, [lessonId, isTeacher, deleting, module_, navigate, setSaveError])

  return { deleting, confirmDelete, setConfirmDelete, requestDeleteLesson, doDeleteLesson }
}
