# Домен: Реакции (лайки и дизлайки)

**Обновлено:** 2026-02-17

## Назначение

Пользователи могут ставить лайки и дизлайки урокам и комментариям. Один пользователь — одна реакция на сущность (повторный запрос перезаписывает: например, с «лайк» на «дизлайк»). Счётчики и своя реакция отдаются в GET урока и GET комментариев.

## Контракт API

- `PUT /api/v1/lessons/:id/reaction` — поставить лайк/дизлайк уроку. Тело: `{ "reaction": "like" | "dislike" }`. JWT.
- `DELETE /api/v1/lessons/:id/reaction` — убрать свою реакцию к уроку. JWT.
- `PUT /api/v1/lessons/:id/comments/:commentId/reaction` — поставить лайк/дизлайк комментарию. Тело: `{ "reaction": "like" | "dislike" }`. JWT.
- `DELETE /api/v1/lessons/:id/comments/:commentId/reaction` — убрать свою реакцию к комментарию. JWT.

Счётчики и `user_reaction` не имеют отдельных эндпоинтов: они подставляются в ответы `GET /api/v1/lessons/:id` и `GET /api/v1/lessons/:id/comments`.

## Связи

- Таблицы `lesson_reactions` (lesson_id, user_id, reaction) и `comment_reactions` (comment_id, user_id, reaction). FK на lessons/users и lesson_comments/users. ON DELETE CASCADE.
- Урок и комментарии домены получают счётчики через провайдеры (LessonReactionProvider, CommentReactionProvider), реализованные в `internal/domain/reactions`.

## Решения

- Реакция — строго `like` или `dislike`. Один пользователь — одна запись на урок/комментарий (UPSERT при PUT).
- Без авторизации счётчики показываются, `user_reaction` не заполняется.
