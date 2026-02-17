import { useTranslation } from 'react-i18next'
import type { LessonStatus } from '../../shared/types'

type LessonProgressBarProps = {
  status: LessonStatus
  onStatusChange: (status: LessonStatus) => void
  isAuthenticated: boolean
}

export function LessonProgressBar({ status, onStatusChange, isAuthenticated }: LessonProgressBarProps) {
  const { t } = useTranslation()
  if (!isAuthenticated) return null

  return (
    <div className="progress-actions">
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
