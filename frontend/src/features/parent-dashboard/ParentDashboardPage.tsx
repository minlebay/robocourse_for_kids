import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '../auth/AuthContext'
import { users as usersApi } from '../../shared/api'
import type { User, UserProgress } from '../../shared/types'

const statusLabels: Record<string, string> = {
  not_started: 'Не начат',
  in_progress: 'В процессе',
  completed: 'Выполнен',
}

export function ParentDashboardPage() {
  const { user: currentUser } = useAuth()
  const [userList, setUserList] = useState<User[]>([])
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deletingId, setDeletingId] = useState<string | null>(null)

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

  const handleDelete = useCallback(
    async (u: User) => {
      if (u.id === currentUser?.id) return
      if (!confirm(`Удалить пользователя ${u.name} (${u.login})?`)) return
      setDeletingId(u.id)
      try {
        await usersApi.delete(u.id)
        setUserList((prev) => prev.filter((x) => x.id !== u.id))
        if (selectedUserId === u.id) setSelectedUserId(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Ошибка удаления')
      } finally {
        setDeletingId(null)
      }
    },
    [currentUser?.id, selectedUserId]
  )

  if (loading) return <p>Загрузка...</p>
  if (error) return <p className="error">{error}</p>

  const selectedUser = userList.find((u) => u.id === selectedUserId)

  return (
    <div className="parent-dashboard">
      <h1>Дашборд родителя</h1>
      <p>Выберите ученика, чтобы посмотреть прогресс.</p>
      <div className="dashboard-layout">
        <aside className="user-list">
          <h2>Пользователи</h2>
          <ul>
            {userList.map((u) => (
              <li key={u.id} className="user-list-item">
                <button
                  className={selectedUserId === u.id ? 'active' : ''}
                  onClick={() => setSelectedUserId(u.id)}
                >
                  {u.name} ({u.login}) {u.role === 'teacher' ? ' [учитель]' : ''}
                </button>
                {u.id !== currentUser?.id && (
                  <button
                    type="button"
                    className="user-delete-btn"
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDelete(u)
                    }}
                    disabled={deletingId === u.id}
                    title="Удалить пользователя"
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
              <h2>Прогресс: {selectedUser.name}</h2>
              {progress === null ? (
                <p>Загрузка прогресса...</p>
              ) : (
                <div>
                  <h3>Уроки</h3>
                  {progress.lessons.length === 0 ? (
                    <p>Нет данных о прогрессе по урокам.</p>
                  ) : (
                    <ul>
                      {progress.lessons.map((l) => (
                        <li key={l.lesson_id}>
                          {l.lesson_title || `Урок ${l.lesson_id.slice(0, 8)}…`} — {statusLabels[l.status] ?? l.status}
                        </li>
                      ))}
                    </ul>
                  )}
                  <h3>Чек-листы</h3>
                  {!progress.checklist || progress.checklist.length === 0 ? (
                    <p>Нет отмеченных пунктов.</p>
                  ) : (
                    <p>Отмечено пунктов: {progress.checklist.length}</p>
                  )}
                </div>
              )}
            </>
          )}
          {!selectedUserId && userList.length > 0 && (
            <p>Выберите пользователя слева.</p>
          )}
        </main>
      </div>
    </div>
  )
}
