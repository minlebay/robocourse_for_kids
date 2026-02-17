import { useTranslation } from 'react-i18next'
import type { LessonStatus } from '../../shared/types'

type LessonProgressBarProps = {
  status: LessonStatus
  onStatusChange: (status: LessonStatus) => void
  isAuthenticated: boolean
  className?: string
}

export function LessonProgressBar({ status, onStatusChange, isAuthenticated, className }: LessonProgressBarProps) {
  const { t } = useTranslation()
  if (!isAuthenticated) return null

  return (
    <div className={className ?? 'progress-actions'} role="group" aria-label={t('progress.statusLabel')}>
      <span>{t('progress.statusLabel')}: </span>
      <button
        type="button"
        className={status === 'not_started' ? 'active' : ''}
        onClick={() => onStatusChange('not_started')}
      >
        {t('progress.status_not_started')}
      </button>
      <button
        type="button"
        className={status === 'in_progress' ? 'active' : ''}
        onClick={() => onStatusChange('in_progress')}
      >
        {t('progress.status_in_progress')}
      </button>
      <button
        type="button"
        className={status === 'completed' ? 'active' : ''}
        onClick={() => onStatusChange('completed')}
      >
        {t('progress.status_completed')}
      </button>
    </div>
  )
}
