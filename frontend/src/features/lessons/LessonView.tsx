import { memo } from 'react'
import { useTranslation } from 'react-i18next'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'
import remarkGfm from 'remark-gfm'
import type { Lesson, LessonStatus } from '../../shared/types'
import { markdownComponents } from './YouTubeEmbed'
import { LessonProgressBar } from './LessonProgressBar'
import { LessonChat } from './LessonChat'
import { LessonComments } from './LessonComments'
import styles from './LessonPage.module.css'

type LessonViewProps = {
  lesson: Lesson
  status: LessonStatus
  isAuthenticated: boolean
  isChecklistDone: (itemId: string) => boolean
  onStatusChange: (status: LessonStatus) => void
  onChecklistToggle: (itemId: string, completed: boolean) => void
  onReactionLike: () => void
  onReactionDislike: () => void
  onClearReaction: () => void
}

function LessonViewInner({
  lesson,
  status,
  isAuthenticated,
  isChecklistDone,
  onStatusChange,
  onChecklistToggle,
  onReactionLike,
  onReactionDislike,
  onClearReaction,
}: LessonViewProps) {
  const { t } = useTranslation()

  return (
    <>
      <h1 className={styles.pageTitle}>{lesson.title}</h1>
      {lesson.description && <p className={styles.description}>{lesson.description}</p>}
      <p className={styles.meta}>
        {t('lesson.typeLabel')}:{' '}
        {t(`module.type${lesson.lesson_type.charAt(0).toUpperCase() + lesson.lesson_type.slice(1)}` as 'module.typeTheory')}
      </p>
      {lesson.tags && lesson.tags.length > 0 && (
        <p className={styles.tags}>{t('lesson.tags')}: {lesson.tags.join(', ')}</p>
      )}

      <LessonProgressBar
        status={status}
        onStatusChange={onStatusChange}
        isAuthenticated={isAuthenticated}
        className={styles.progressActions}
      />

      {lesson.steps && lesson.steps.length > 0 && (
        <section className={styles.sectionSteps} aria-labelledby="lesson-steps-heading">
          <h2 id="lesson-steps-heading">{t('lesson.steps')}</h2>
          {[...lesson.steps]
            .sort((a, b) => a.sort_order - b.sort_order)
            .map((step) => (
              <article key={step.id} className={styles.step}>
                <h3>{step.title}</h3>
                <div className={`${styles.stepContent} ${styles.stepContentMarkdown}`}>
                  <ReactMarkdown rehypePlugins={[rehypeHighlight]} remarkPlugins={[remarkGfm]} components={markdownComponents}>
                    {step.content ?? ''}
                  </ReactMarkdown>
                </div>
              </article>
            ))}
        </section>
      )}

      {lesson.checklist && lesson.checklist.length > 0 && (
        <section className={styles.sectionChecklist} aria-labelledby="lesson-checklist-heading">
          <h2 id="lesson-checklist-heading">{t('lesson.checklist')}</h2>
          <ul className={styles.checklist}>
            {[...lesson.checklist]
              .sort((a, b) => a.sort_order - b.sort_order)
              .map((item) => (
                <li key={item.id}>
                  <label>
                    <input
                      type="checkbox"
                      checked={isChecklistDone(item.id)}
                      onChange={(e) => onChecklistToggle(item.id, e.target.checked)}
                      disabled={!isAuthenticated}
                      aria-label={item.title}
                    />
                    {item.title}
                  </label>
                </li>
              ))}
          </ul>
        </section>
      )}

      <div className={styles.reactions} role="group" aria-label={t('lesson.reactions')}>
        <button
          type="button"
          className={`${styles.reactionBtn} ${lesson.user_reaction === 'like' ? styles.reactionBtnActive : ''}`}
          onClick={lesson.user_reaction === 'like' ? onClearReaction : onReactionLike}
          title={t('comments.like')}
          aria-pressed={lesson.user_reaction === 'like'}
          disabled={!isAuthenticated}
        >
          👍 <span className={styles.reactionCount}>{lesson.likes_count ?? 0}</span>
        </button>
        <button
          type="button"
          className={`${styles.reactionBtn} ${lesson.user_reaction === 'dislike' ? styles.reactionBtnActive : ''}`}
          onClick={lesson.user_reaction === 'dislike' ? onClearReaction : onReactionDislike}
          title={t('comments.dislike')}
          aria-pressed={lesson.user_reaction === 'dislike'}
          disabled={!isAuthenticated}
        >
          👎 <span className={styles.reactionCount}>{lesson.dislikes_count ?? 0}</span>
        </button>
      </div>

      <LessonChat lessonId={lesson.id} />
      <LessonComments lessonId={lesson.id} />
    </>
  )
}

export const LessonView = memo(LessonViewInner)
