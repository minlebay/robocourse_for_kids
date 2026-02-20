import { useTranslation } from 'react-i18next'
import type { AdminStats } from '../../shared/types'

interface StatsCardsProps {
  stats: AdminStats | null
  loading: boolean
}

export function StatsCards({ stats, loading }: StatsCardsProps) {
  const { t } = useTranslation()

  const cards = [
    { label: t('admin.users'), value: stats?.users },
    { label: t('admin.courses'), value: stats?.modules },
    { label: t('admin.lessons'), value: stats?.lessons },
  ]

  return (
    <div className="stats-grid">
      {cards.map((card) => (
        <div key={card.label} className="stat-card">
          <div className="stat-card-label">{card.label}</div>
          <div className="stat-card-value">
            {loading ? '—' : (card.value ?? '—')}
          </div>
        </div>
      ))}
    </div>
  )
}
