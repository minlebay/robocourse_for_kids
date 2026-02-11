# Домен: Прогресс по урокам и чек-листам

**Обновлено:** 2025-02-09

## Назначение

Отметки пользователя: статус по уроку (not_started | in_progress | completed) и выполнение пунктов чек-листа. Привязаны к `user_id` и `lesson_id` / `checklist_item_id`.

## Контракт API

- `GET /api/v1/progress` — прогресс текущего пользователя (требуется авторизация). В каждом элементе lessons: lesson_id, lesson_title, module_id, status, updated_at (для подсказки «Продолжить» в шапке).
- `PUT /api/v1/lessons/:id/progress` — тело: `{ "status": "not_started" | "in_progress" | "completed" }`.
- `PUT /api/v1/lessons/:id/checklist/:itemId` — тело: `{ "completed": true|false }` (по умолчанию true).

## Связи

- Зависит от доменов users и lessons (lesson_id, checklist_item_id из lessons).
- Родительский дашборд использует тот же репозиторий прогресса для запроса по чужому user_id (доступно только teacher).

## Решения

- Первичный ключ прогресса по уроку: (user_id, lesson_id). Обновление через INSERT ... ON CONFLICT DO UPDATE.
- Чек-лист: отметка хранится в `user_checklist_progress` (user_id, checklist_item_id, completed_at).
