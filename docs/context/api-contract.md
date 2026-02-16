# Сводка контракта API

**Обновлено:** 2026-02-16

Базовый префикс: `/api/v1`. Полная спецификация: [backend/api/openapi.yaml](../../backend/api/openapi.yaml).

## Эндпоинты

| Метод | Путь | Описание | Auth | Rate limit |
|-------|------|----------|------|------------|
| GET | /api/v1/health | Готовность и БД | — | — |
| POST | /api/v1/auth/register | Регистрация | — | 10/мин |
| POST | /api/v1/auth/login | Вход | — | 10/мин |
| GET | /api/v1/auth/me | Текущий пользователь | JWT | — |
| PATCH | /api/v1/auth/me | Обновить профиль (тема) | JWT | — |
| GET | /api/v1/modules | Список модулей (?tag=...) | — | — |
| POST | /api/v1/modules | Создать курс | JWT, teacher | — |
| DELETE | /api/v1/modules/:id | Удалить курс | JWT, teacher | — |
| GET | /api/v1/modules/:id | Модуль с уроками | — | — |
| POST | /api/v1/modules/:id/lessons | Добавить урок в курс | JWT, teacher | — |
| GET | /api/v1/lessons/:id | Урок (шаги, материалы, чек-лист) | — | — |
| PUT | /api/v1/lessons/:id | Обновить урок | JWT, teacher | — |
| DELETE | /api/v1/lessons/:id | Удалить урок | JWT, teacher | — |
| GET | /api/v1/lessons/:id/comments | Список комментариев к уроку | — | — |
| POST | /api/v1/lessons/:id/comments | Добавить комментарий | JWT | — |
| DELETE | /api/v1/lessons/:id/comments/:commentId | Удалить свой комментарий | JWT | — |
| GET | /api/v1/progress | Прогресс текущего пользователя | JWT | — |
| PUT | /api/v1/lessons/:id/progress | Статус по уроку | JWT | — |
| PUT | /api/v1/lessons/:id/checklist/:itemId | Пункт чек-листа | JWT | — |
| POST | /api/v1/chat | Отправить сообщение в Gemini | JWT | 20/мин |
| GET | /api/v1/chat/:lessonId/history | История чата по уроку | JWT | — |
| DELETE | /api/v1/chat/:lessonId/history | Очистить историю чата | JWT | — |
| GET | /api/v1/users | Список пользователей | JWT, teacher | — |
| DELETE | /api/v1/users/:id | Удалить пользователя (не себя) | JWT, teacher | — |
| GET | /api/v1/users/:id/progress | Прогресс ученика (404 если пользователь не найден) | JWT, teacher | — |

## Форматы

- Ответы — JSON. Коды: 200, 201, 204, 400, 401, 403, 404, 429, 500, 502, 503.
- Ошибки: тело `{ "error": "..." }`.
- JWT: заголовок `Authorization: Bearer <token>`. Время жизни: 1 час.
- Max request body: 1 MB.
- Невалидный Bearer → 401 (ранее молча пропускался).

## Безопасность

- Регистрация teacher требует `invite_code` (env `TEACHER_INVITE_CODE`).
- Пароль: 6–72 символа. Login: 3–50 символов. Имя: до 200 символов.
- Rate limiting: auth 10/мин, chat 20/мин (per IP).
- System prompt для чата формируется на сервере (поле `lesson_context` удалено из запроса).
- Chat: клиент отправляет только `message` (строка), историю сервер загружает из БД.
- Structured logging (slog JSON) для production observability.
- Репозитории определены через интерфейсы для тестируемости.
- Контент шагов уроков (markdown) и описания санитизируются на бэкенде (bluemonday) перед записью; лимиты длины: title 500, description 10 KB, step content 100 KB, steps per lesson 200. При создании урока `lesson_type`: только `theory`, `practice` или `project` (400 при неверном).
- Текст комментария санитизируется (bluemonday) перед записью; длина 1–2000 символов.
