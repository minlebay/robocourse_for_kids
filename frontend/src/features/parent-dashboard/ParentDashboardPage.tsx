import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../auth/AuthContext'
import { users as usersApi } from '../../shared/api'
import type { User, UserProgress } from '../../shared/types'
import { ConfirmModal } from '../../components'

function statusLabel(t: (key: string) => string, status: string): string {
  const key = `progress.status_${status}` as 'progress.status_not_started'
  return t(key)
}

export function ParentDashboardPage() {
  const { user: currentUser } = useAuth()
  const [userList, setUserList] = useState<User[]>([])
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [confirmUser, setConfirmUser] = useState<User | null>(null)
  const { t } = useTranslation()

  const loadUsers = useCallback(() => {
    usersApi
      .list()
      .then(setUserList)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    loadUsers()
  }, [loadUsers])

  useEffect(() => {
    if (!selectedUserId) {
      setProgress(null)
      return
    }
    setProgress(null)
    usersApi
      .progress(selectedUserId)
      .then(setProgress)
      .catch(() => setProgress(null))
  }, [selectedUserId])

  const requestDeleteUser = useCallback((u: User) => {
    if (u.id === currentUser?.id) return
    setConfirmUser(u)
  }, [currentUser?.id])

  const doDeleteUser = useCallback(
    async () => {
      const u = confirmUser
      if (!u) return
      setConfirmUser(null)
      setDeletingId(u.id)
      try {
        await usersApi.delete(u.id)
        setUserList((prev) => prev.filter((x) => x.id !== u.id))
        if (selectedUserId === u.id) setSelectedUserId(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : t('dashboard.deleteUserError'))
      } finally {
        setDeletingId(null)
      }
    },
    [confirmUser, selectedUserId, t]
  )

  if (loading) return <p>{t('common.loading')}</p>
  if (error) return <p className="error">{error}</p>

  const selectedUser = userList.find((u) => u.id === selectedUserId)

  return (
    <div className="parent-dashboard">
      <ConfirmModal
        open={!!confirmUser}
        title={t('dashboard.confirmDeleteUser')}
        confirmLabel={t('common.delete')}
        variant="danger"
        onConfirm={doDeleteUser}
        onCancel={() => setConfirmUser(null)}
      >
        {confirmUser
          ? t('dashboard.confirmDeleteUserMessage', { name: confirmUser.name, login: confirmUser.login })
          : ''}
      </ConfirmModal>
      <h1>{t('dashboard.title')}</h1>
      <p>{t('dashboard.selectStudent')}</p>
      <div className="dashboard-layout">
        <aside className="user-list">
          <h2>{t('dashboard.users')}</h2>
          <ul>
            {userList.map((u) => (
              <li key={u.id} className="user-list-item">
                <button
                  className={selectedUserId === u.id ? 'active' : ''}
                  onClick={() => setSelectedUserId(u.id)}
                >
                  {u.name} ({u.login}) {u.role === 'teacher' ? t('dashboard.teacherBadge') : ''}
                </button>
                {u.id !== currentUser?.id && (
                  <button
                    type="button"
                    className="user-delete-btn"
                    onClick={(e) => {
                      e.stopPropagation()
                      requestDeleteUser(u)
                    }}
                    disabled={deletingId === u.id}
                    title={t('dashboard.deleteUser')}
                  >
                    {deletingId === u.id ? '…' : '✕'}
                  </button>
                )}
              </li>
            ))}
          </ul>
        </aside>
        <main className="progress-panel">
          {selectedUser && (
            <>
              <h2>{t('dashboard.progressLabel', { name: selectedUser.name })}</h2>
              {progress === null ? (
                <p>{t('dashboard.loadingProgress')}</p>
              ) : (
                <div>
                  <h3>{t('dashboard.lessons')}</h3>
                  {progress.lessons.length === 0 ? (
                    <p>{t('dashboard.noLessonsProgress')}</p>
                  ) : (
                    <ul>
                      {progress.lessons.map((l) => (
                        <li key={l.lesson_id}>
                          {l.lesson_title || t('progress.lessonFallback', { id: l.lesson_id.slice(0, 8) })} — {statusLabel(t, l.status)}
                        </li>
                      ))}
                    </ul>
                  )}
                  <h3>{t('dashboard.checklists')}</h3>
                  {!progress.checklist || progress.checklist.length === 0 ? (
                    <p>{t('dashboard.noChecklistItems')}</p>
                  ) : (
                    <p>{t('dashboard.checklistCount', { count: progress.checklist.length })}</p>
                  )}
                </div>
              )}
            </>
          )}
          {!selectedUserId && userList.length > 0 && (
            <p>{t('dashboard.selectUser')}</p>
          )}
        </main>
      </div>
    </div>
  )
}
