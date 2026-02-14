import type { LessonStatus } from '../../shared/types'

type LessonProgressBarProps = {
  status: LessonStatus
  onStatusChange: (status: LessonStatus) => void
  isAuthenticated: boolean
}

export function LessonProgressBar({ status, onStatusChange, isAuthenticated }: LessonProgressBarProps) {
  if (!isAuthenticated) return null

  return (
    <div className="progress-actions">
      <span>Статус: </span>
      <button
        type="button"
        className={status === 'not_started' ? 'active' : ''}
        onClick={() => onStatusChange('not_started')}
      >
        Не начат
      </button>
      <button
        type="button"
        className={status === 'in_progress' ? 'active' : ''}
        onClick={() => onStatusChange('in_progress')}
      >
        В процессе
      </button>
      <button
        type="button"
        className={status === 'completed' ? 'active' : ''}
        onClick={() => onStatusChange('completed')}
      >
        Выполнен
      </button>
    </div>
  )
}
