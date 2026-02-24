import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { admin as adminApi } from '../../shared/api'
import type { AdminStats, ActivityItem, User, AdminCreateUserRequest } from '../../shared/types'
import { useAuth } from '../auth/AuthContext'
import { ErrorBoundary } from '../../shared/ErrorBoundary'
import { StatsCards } from './StatsCards'
import { ActivityFeed } from './ActivityFeed'
import { TempPasswordModal } from './TempPasswordModal'
import { CreateUserModal } from './CreateUserModal'
import { UserTable } from './UserTable'
import './AdminPage.css'

export function AdminPage() {
  const { t } = useTranslation()
  const { user } = useAuth()

  const [stats, setStats] = useState<AdminStats | null>(null)
  const [statsLoading, setStatsLoading] = useState(true)

  const [activity, setActivity] = useState<ActivityItem[]>([])
  const [activityLoading, setActivityLoading] = useState(true)

  const [users, setUsers] = useState<User[]>([])
  const [usersLoading, setUsersLoading] = useState(true)

  const [showCreateModal, setShowCreateModal] = useState(false)
  const [tempPassword, setTempPassword] = useState<string | null>(null)

  const [actionError, setActionError] = useState('')
  const [loadError, setLoadError] = useState('')

  const loadUsers = useCallback(async () => {
    setUsersLoading(true)
    try {
      const data = await adminApi.listUsers()
      setUsers(data)
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : t('errors.loadFailed'))
    } finally {
      setUsersLoading(false)
    }
  }, [t])

  const loadStats = useCallback(async () => {
    try {
      const data = await adminApi.getStats()
      setStats(data)
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : t('errors.loadFailed'))
    } finally {
      setStatsLoading(false)
    }
  }, [t])

  const loadActivity = useCallback(async () => {
    try {
      const data = await adminApi.getActivity()
      setActivity(data)
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : t('errors.loadFailed'))
    } finally {
      setActivityLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadStats()
    loadActivity()
    loadUsers()
  }, [loadStats, loadActivity, loadUsers])

  async function handleBlock(id: string, block: boolean) {
    setActionError('')
    try {
      await adminApi.blockUser(id, block)
      await loadUsers()
    } catch (err) {
      setActionError(err instanceof Error ? err.message : t('errors.somethingWrong'))
    }
  }

  async function handleDelete(id: string) {
    setActionError('')
    try {
      await adminApi.deleteUser(id)
      await loadUsers()
    } catch (err) {
      setActionError(err instanceof Error ? err.message : t('errors.somethingWrong'))
    }
  }

  async function handleResetPassword(id: string) {
    setActionError('')
    try {
      const { temp_password } = await adminApi.resetPassword(id)
      setTempPassword(temp_password)
    } catch (err) {
      setActionError(err instanceof Error ? err.message : t('errors.somethingWrong'))
    }
  }

  async function handleCreateUser(data: AdminCreateUserRequest): Promise<{ user: User; temp_password: string }> {
    const result = await adminApi.createUser(data)
    setShowCreateModal(false)
    setTempPassword(result.temp_password)
    await loadUsers()
    return result
  }

  return (
    <ErrorBoundary>
      <div className="admin-page">
        <h1>{t('admin.title')}</h1>

        <StatsCards stats={stats} loading={statsLoading} />

        {loadError && (
          <p className="error" role="alert" style={{ marginBottom: '1rem' }}>
            {loadError}
            <button
              type="button"
              className="button-secondary"
              style={{ marginLeft: '1rem', padding: '0.2rem 0.75rem', fontSize: '0.85rem' }}
              onClick={() => { setLoadError(''); loadStats(); loadActivity(); loadUsers() }}
            >
              {t('common.retry')}
            </button>
          </p>
        )}

        {actionError && (
          <p className="error" role="alert" style={{ marginBottom: '1rem' }}>
            {actionError}
          </p>
        )}

        <div className="admin-main">
          <div className="admin-users-panel">
            <div className="admin-header">
              <h2>{t('admin.usersTitle')}</h2>
              <button
                type="button"
                className="button-primary"
                onClick={() => setShowCreateModal(true)}
              >
                + {t('admin.createUser')}
              </button>
            </div>
            <UserTable
              users={users}
              loading={usersLoading}
              currentUserId={user?.id ?? ''}
              onBlock={handleBlock}
              onDelete={handleDelete}
              onResetPassword={handleResetPassword}
            />
          </div>

          <ActivityFeed items={activity} loading={activityLoading} />
        </div>

        {showCreateModal && (
          <CreateUserModal
            onClose={() => setShowCreateModal(false)}
            onCreate={handleCreateUser}
          />
        )}

        {tempPassword !== null && (
          <TempPasswordModal
            password={tempPassword}
            onClose={() => setTempPassword(null)}
          />
        )}
      </div>
    </ErrorBoundary>
  )
}
