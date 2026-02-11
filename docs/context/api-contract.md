# Сводка контракта API

**Обновлено:** 2025-02-09

Базовый префикс: `/api/v1`. Полная спецификация: [backend/api/openapi.yaml](../../backend/api/openapi.yaml).

## Эндпоинты

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /api/v1/health | Готовность и БД | — |
| POST | /api/v1/auth/register | Регистрация | — |
| POST | /api/v1/auth/login | Вход | — |
| GET | /api/v1/auth/me | Текущий пользователь | JWT |
| GET | /api/v1/modules | Список модулей | — |
| GET | /api/v1/modules/:id | Модуль с уроками | — |
| GET | /api/v1/lessons/:id | Урок (шаги, материалы, чек-лист) | — |
| GET | /api/v1/lessons/:id/comments | Список комментариев к уроку | — |
| POST | /api/v1/lessons/:id/comments | Добавить комментарий | JWT |
| DELETE | /api/v1/lessons/:id/comments/:commentId | Удалить свой комментарий | JWT |
| GET | /api/v1/progress | Прогресс текущего пользователя | JWT |
| PUT | /api/v1/lessons/:id/progress | Статус по уроку | JWT |
| PUT | /api/v1/lessons/:id/checklist/:itemId | Пункт чек-листа | JWT |
| GET | /api/v1/users | Список пользователей | JWT, teacher |
| GET | /api/v1/users/:id/progress | Прогресс ученика | JWT, teacher |

## Форматы

- Ответы — JSON. Коды: 200, 201, 400, 401, 403, 404, 500.
- Ошибки: тело `{ "error": "..." }` или `{ "message": "..." }`.
- JWT: заголовок `Authorization: Bearer <token>`.
