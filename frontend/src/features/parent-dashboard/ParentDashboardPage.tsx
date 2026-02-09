import { useCallback, useEffect, useState } from 'react'
import { users as usersApi } from '../../shared/api'
import type { User, UserProgress } from '../../shared/types'

export function ParentDashboardPage() {
  const [userList, setUserList] = useState<User[]>([])
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)
  const [progress, setProgress] = useState<UserProgress | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

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
              <li key={u.id}>
                <button
                  className={selectedUserId === u.id ? 'active' : ''}
                  onClick={() => setSelectedUserId(u.id)}
                >
                  {u.name} ({u.login}) {u.role === 'teacher' ? ' [учитель]' : ''}
                </button>
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
                          Урок {l.lesson_id.slice(0, 8)}… — {l.status}
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
