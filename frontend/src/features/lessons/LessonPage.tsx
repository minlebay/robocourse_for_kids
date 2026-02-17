import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../auth/AuthContext'
import { PageWithToc, type TocItem } from './PageWithToc'
import { ConfirmModal } from '../../components'
import { CourseNav } from './CourseNav'
import { LessonView } from './LessonView'
import { LessonEditForm } from './LessonEditForm'
import { useLessonData, useLessonActions, useLessonEdit, useLessonDelete } from './hooks'
import styles from './LessonPage.module.css'

export function LessonPage() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const { t } = useTranslation()

  const { lesson, module_, progress, loading, error, load, setLesson, setProgress, setError } = useLessonData(
    id,
    !!user
  )

  const {
    setLessonStatus,
    toggleChecklist,
    setLessonReaction,
    clearLessonReaction,
    getLessonStatus,
    isChecklistDone,
  } = useLessonActions({
    lessonId: id,
    isAuthenticated: !!user,
    progress,
    load,
    setError,
    setLesson,
    setProgress,
  })

  const {
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
  } = useLessonEdit({ lessonId: id, lesson, setLesson })

  const { deleting, confirmDelete, setConfirmDelete, requestDeleteLesson, doDeleteLesson } = useLessonDelete({
    lessonId: id,
    isTeacher: user?.role === 'teacher',
    module_,
    setSaveError,
  })

  if (loading) return <p>{t('common.loading')}</p>
  if (error) return <p className="error">{error}</p>
  if (!lesson) return <p className="error">{t('lesson.notFound')}</p>

  const status = getLessonStatus()
  const orderedLessons = (module_?.lessons ?? []).sort((a, b) => a.sort_order - b.sort_order)
  const tocItems: TocItem[] = orderedLessons.map((l) => ({
    id: l.id,
    title: l.title,
    href: `/lessons/${l.id}`,
    active: l.id === lesson.id,
  }))
  const tocTitle = module_ ? module_.title : t('lesson.tocTitle')

  const lessonContent = (
    <div
      className={styles.root}
      data-lesson-page
    >
      <CourseNav module={module_} currentLesson={lesson} className={styles.courseNav} />
      <div className={styles.content} data-lesson-page-content>
        {user?.role === 'teacher' && !editing && (
          <div className={styles.editBar}>
            <button type="button" className="button-primary" onClick={startEditing}>
              {t('lesson.editLesson')}
            </button>
            <button
              type="button"
              className="button-danger-outline"
              onClick={requestDeleteLesson}
              disabled={deleting}
              title={t('lesson.deleteLesson')}
              aria-label={t('lesson.deleteLesson')}
            >
              {deleting ? '…' : t('lesson.deleteLesson')}
            </button>
          </div>
        )}

        {editing && editForm ? (
          <LessonEditForm
            form={editForm}
            saveError={saveError}
            onTitleChange={setFormTitle}
            onDescriptionChange={setFormDescription}
            onStepTitleChange={(stepId, value) => updateEditStep(stepId, 'title', value)}
            onStepContentChange={(stepId, value) => updateEditStep(stepId, 'content', value)}
            onRemoveStep={removeStep}
            onAddStep={addStep}
            onSave={saveLesson}
            onCancel={cancelEditing}
          />
        ) : (
          <>
            <LessonView
              lesson={lesson}
              status={status}
              isAuthenticated={!!user}
              isChecklistDone={isChecklistDone}
              onStatusChange={setLessonStatus}
              onChecklistToggle={toggleChecklist}
              onReactionLike={() => setLessonReaction('like')}
              onReactionDislike={() => setLessonReaction('dislike')}
              onClearReaction={clearLessonReaction}
            />
          </>
        )}
      </div>
      <CourseNav module={module_} currentLesson={lesson} className={styles.courseNavLast} />
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
        title={t('lesson.confirmDelete')}
        confirmLabel={t('common.delete')}
        variant="danger"
        onConfirm={doDeleteLesson}
        onCancel={() => setConfirmDelete(false)}
      >
        {lesson ? t('lesson.confirmDeleteMessage', { title: lesson.title }) : ''}
      </ConfirmModal>
      {lessonContent}
    </PageWithToc>
  )
}
