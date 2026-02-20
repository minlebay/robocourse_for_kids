import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ConfirmModal } from '../../components/ConfirmModal'
import type { User } from '../../shared/types'

interface UserTableProps {
  users: User[]
  loading: boolean
  currentUserId: string
  onBlock: (id: string, block: boolean) => void
  onDelete: (id: string) => void
  onResetPassword: (id: string) => void
}

const ROLE_BADGE_COLORS: Record<string, { bg: string; color: string }> = {
  student: { bg: '#1e3a5f', color: '#93c5fd' },
  teacher: { bg: '#064e3b', color: '#6ee7b7' },
  administrator: { bg: '#3b0764', color: '#d8b4fe' },
}

function RoleBadge({ role }: { role: string }) {
  const colors = ROLE_BADGE_COLORS[role] ?? { bg: '#334155', color: '#94a3b8' }
  return (
    <span
      className="role-badge"
      style={{ backgroundColor: colors.bg, color: colors.color }}
    >
      {role}
    </span>
  )
}

function StatusDot({ blocked }: { blocked: boolean }) {
  return (
    <span
      className="status-dot"
      title={blocked ? 'Заблокирован' : 'Активен'}
      style={{ backgroundColor: blocked ? '#ef4444' : '#10b981' }}
    />
  )
}

export function UserTable({ users, loading, currentUserId, onBlock, onDelete, onResetPassword }: UserTableProps) {
  const { t } = useTranslation()
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null)

  if (loading) {
    return <p style={{ color: 'var(--text-muted)' }}>{t('common.loading')}</p>
  }

  return (
    <>
      <table className="user-table">
        <thead>
          <tr>
            <th>{t('admin.loginCol')}</th>
            <th>{t('admin.nameCol')}</th>
            <th>{t('admin.roleCol')}</th>
            <th>{t('admin.emailCol')}</th>
            <th>{t('admin.statusCol')}</th>
            <th>{t('admin.actionsCol')}</th>
          </tr>
        </thead>
        <tbody>
          {users.map((u) => {
            const isSelf = u.id === currentUserId
            const isBlocked = u.is_blocked === true
            return (
              <tr key={u.id} className={isBlocked ? 'user-row-blocked' : ''}>
                <td className="user-login">@{u.login}</td>
                <td>{u.name}</td>
                <td><RoleBadge role={u.role} /></td>
                <td className="user-email">{u.email ?? '—'}</td>
                <td>
                  <StatusDot blocked={isBlocked} />
                  <span className="status-label">
                    {isBlocked ? t('admin.statusBlocked') : t('admin.statusActive')}
                  </span>
                </td>
                <td className="user-actions">
                  <button
                    type="button"
                    className="action-btn"
                    title={isBlocked ? t('admin.unblock') : t('admin.block')}
                    disabled={isSelf}
                    onClick={() => onBlock(u.id, !isBlocked)}
                  >
                    {isBlocked ? '🔓' : '🚫'}
                  </button>
                  <button
                    type="button"
                    className="action-btn"
                    title={t('admin.resetPassword')}
                    onClick={() => onResetPassword(u.id)}
                  >
                    🔑
                  </button>
                  <button
                    type="button"
                    className="action-btn action-btn-danger"
                    title={t('admin.delete')}
                    disabled={isSelf}
                    onClick={() => setDeleteTarget(u)}
                  >
                    🗑
                  </button>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>

      <ConfirmModal
        open={deleteTarget !== null}
        title={t('admin.confirmDelete')}
        confirmLabel={t('common.delete')}
        variant="danger"
        onConfirm={() => {
          if (deleteTarget) {
            onDelete(deleteTarget.id)
            setDeleteTarget(null)
          }
        }}
        onCancel={() => setDeleteTarget(null)}
      >
        <p>{t('admin.confirmDeleteDesc')}</p>
        {deleteTarget && (
          <p>
            <strong>@{deleteTarget.login}</strong> — {deleteTarget.name}
          </p>
        )}
      </ConfirmModal>
    </>
  )
}
