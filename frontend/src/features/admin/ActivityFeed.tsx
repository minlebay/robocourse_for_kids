import { useTranslation } from 'react-i18next'
import type { ActivityItem } from '../../shared/types'

interface ActivityFeedProps {
  items: ActivityItem[]
  loading: boolean
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMinutes = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMinutes < 1) return 'только что'
  if (diffMinutes < 60) return `${diffMinutes} мин назад`
  if (diffHours < 24) return `${diffHours} ч назад`
  if (diffDays < 7) return `${diffDays} дн назад`
  return date.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

export function ActivityFeed({ items, loading }: ActivityFeedProps) {
  const { t } = useTranslation()

  return (
    <div className="activity-feed">
      <h2 className="activity-title">{t('admin.activityTitle')}</h2>
      {loading && <p style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</p>}
      {!loading && items.length === 0 && (
        <p style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>—</p>
      )}
      <ul className="activity-list">
        {items.map((item) => (
          <li key={item.id} className="activity-item">
            <div className="activity-name">{item.name}</div>
            <div className="activity-meta">
              <span className="activity-login">@{item.login}</span>
              <span className="activity-sep">·</span>
              <span className="activity-time">{formatRelativeTime(item.created_at)}</span>
            </div>
            <div className="activity-action">{t('admin.activityRegistered')}</div>
          </li>
        ))}
      </ul>
    </div>
  )
}
