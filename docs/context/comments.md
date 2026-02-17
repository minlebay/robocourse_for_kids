# Домен: Комментарии к урокам

**Обновлено:** 2026-02-17

## Назначение

Пользователи (ученики и учителя) могут оставлять комментарии к урокам. Комментарии видны всем, добавлять можно только авторизованным пользователям.

## Контракт API

- `GET /api/v1/lessons/:id/comments` — список комментариев к уроку (без авторизации). В каждом комментарии есть `likes_count`, `dislikes_count`, `user_reaction` (если пользователь авторизован).
- `POST /api/v1/lessons/:id/comments` — добавить комментарий. Тело: `{ "text": "..." }`. Требуется JWT.
- `DELETE /api/v1/lessons/:id/comments/:commentId` — удалить свой комментарий. Требуется JWT. Удалить можно только свой комментарий.
- `PUT /api/v1/lessons/:id/comments/:commentId/reaction` — поставить лайк/дизлайк комментарию. Тело: `{ "reaction": "like" | "dislike" }`. JWT.
- `DELETE /api/v1/lessons/:id/comments/:commentId/reaction` — убрать свою реакцию к комментарию. JWT.

## Формат комментария

```json
{
  "id": "uuid",
  "lesson_id": "uuid",
  "user_id": "uuid",
  "user_name": "Имя пользователя",
  "text": "Текст комментария",
  "created_at": "2025-02-11T12:00:00Z",
  "likes_count": 0,
  "dislikes_count": 0,
  "user_reaction": "like"
}
```

- `likes_count`, `dislikes_count` — число лайков и дизлайков. `user_reaction` — реакция текущего пользователя (`"like"` или `"dislike"`), присутствует только при авторизации.

- `text` — 1–2000 символов. Перед записью в БД текст санитизируется (bluemonday) для защиты от XSS.

## Связи

- Таблица `lesson_comments`: `lesson_id` → lessons, `user_id` → users.
- При удалении урока или пользователя комментарии каскадно удаляются.
