import { useCallback, useEffect, useState } from 'react'
import { lessons as lessonsApi } from '../../shared/api'
import { useAuth } from '../auth/AuthContext'
import { COMMENT_MAX } from '../../shared/validation'
import type { LessonComment } from '../../shared/types'

function formatDate(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return iso
  }
}

export function LessonComments({ lessonId }: { lessonId: string }) {
  const { user } = useAuth()
  const [comments, setComments] = useState<LessonComment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [text, setText] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const load = useCallback(() => {
    lessonsApi
      .getComments(lessonId)
      .then(setComments)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [lessonId])

  useEffect(() => {
    load()
  }, [load])

  const submit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      const trimmed = text.trim()
      if (!trimmed || !user || submitting) return
      if (trimmed.length > COMMENT_MAX) {
        setError(`Комментарий не более ${COMMENT_MAX} символов`)
        return
      }
      setError('')
      setSubmitting(true)
      lessonsApi
        .addComment(lessonId, trimmed)
        .then((newComment) => {
          setComments((prev) => [...prev, newComment])
          setText('')
        })
        .catch((err) => setError(err.message))
        .finally(() => setSubmitting(false))
    },
    [lessonId, text, user, submitting]
  )

  const handleDelete = useCallback(
    (commentId: string) => {
      if (!user || deletingId) return
      setDeletingId(commentId)
      setError('')
      lessonsApi
        .deleteComment(lessonId, commentId)
        .then(() => setComments((prev) => prev.filter((c) => c.id !== commentId)))
        .catch((err) => setError(err.message))
        .finally(() => setDeletingId(null))
    },
    [lessonId, user, deletingId]
  )

  const setCommentReaction = useCallback(
    (commentId: string, reaction: import('../../shared/types').ReactionType) => {
      if (!user) return
      const comment = comments.find((c) => c.id === commentId)
      if (comment?.user_reaction === reaction) {
        lessonsApi.deleteCommentReaction(lessonId, commentId).then(load).catch((err) => setError(err.message))
      } else {
        lessonsApi.setCommentReaction(lessonId, commentId, reaction).then(load).catch((err) => setError(err.message))
      }
    },
    [lessonId, user, comments, load]
  )

  if (loading) {
    return (
      <section className="lesson-comments">
        <h2>Комментарии</h2>
        <p className="lesson-comments-loading">Загрузка...</p>
      </section>
    )
  }

  return (
    <section className="lesson-comments">
      <h2>Комментарии</h2>
      {error && <p className="lesson-comments-error">{error}</p>}
      <ul className="lesson-comments-list" aria-label="Комментарии к уроку">
        {comments.length === 0 ? (
          <li className="lesson-comments-empty">Пока нет комментариев. Будь первым!</li>
        ) : (
          comments.map((c) => (
            <li key={c.id} className="lesson-comments-item">
              <div className="lesson-comments-meta">
                <span className="lesson-comments-author">{c.user_name}</span>
                <span className="lesson-comments-date">{formatDate(c.created_at)}</span>
                {user?.id === c.user_id && (
                  <button
                    type="button"
                    className="lesson-comments-delete button-danger-outline"
                    onClick={() => handleDelete(c.id)}
                    disabled={deletingId === c.id}
                    title="Удалить комментарий"
                  >
                    {deletingId === c.id ? '…' : 'Удалить'}
                  </button>
                )}
              </div>
              <p className="lesson-comments-text">{c.text}</p>
              <div className="comment-reactions" aria-label="Реакции к комментарию">
                <button
                  type="button"
                  className={`comment-reaction-btn ${c.user_reaction === 'like' ? 'active' : ''}`}
                  onClick={user ? () => setCommentReaction(c.id, 'like') : undefined}
                  title="Нравится"
                  aria-pressed={c.user_reaction === 'like'}
                  disabled={!user}
                >
                  👍 <span className="comment-reaction-count">{c.likes_count ?? 0}</span>
                </button>
                <button
                  type="button"
                  className={`comment-reaction-btn ${c.user_reaction === 'dislike' ? 'active' : ''}`}
                  onClick={user ? () => setCommentReaction(c.id, 'dislike') : undefined}
                  title="Не нравится"
                  aria-pressed={c.user_reaction === 'dislike'}
                  disabled={!user}
                >
                  👎 <span className="comment-reaction-count">{c.dislikes_count ?? 0}</span>
                </button>
              </div>
            </li>
          ))
        )}
      </ul>
      {user ? (
        <form className="lesson-comments-form" onSubmit={submit}>
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Напишите комментарий..."
            rows={2}
            maxLength={COMMENT_MAX}
            className="lesson-comments-input"
            disabled={submitting}
          />
          <div className="lesson-comments-actions">
            <span className="lesson-comments-hint">
              {text.length}/{COMMENT_MAX}
            </span>
            <button type="submit" className="button-primary" disabled={!text.trim() || submitting}>
              Отправить
            </button>
          </div>
        </form>
      ) : (
        <p className="lesson-comments-login">Войдите, чтобы оставить комментарий.</p>
      )}
    </section>
  )
}
