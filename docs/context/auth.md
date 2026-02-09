# Домен: Авторизация (пользователи, роли, JWT)

**Обновлено:** 2025-02-09

## Назначение

Регистрация, вход, выдача JWT; роли `student` и `teacher`. Только учитель имеет доступ к списку пользователей и к прогрессу других учеников.

## Контракт API

- `POST /api/v1/auth/register` — тело: `{ "login", "password", "name", "role?" }`. Ответ: `{ "user", "token" }`.
- `POST /api/v1/auth/login` — тело: `{ "login", "password" }`. Ответ: `{ "user", "token" }`.
- `GET /api/v1/auth/me` — текущий пользователь (требуется заголовок `Authorization: Bearer <token>`).

## Решения

- JWT в заголовке `Authorization: Bearer <token>`.
- Пароли хранятся как bcrypt-хеш.
- Роль по умолчанию при регистрации — `student`. Роль `teacher` даёт доступ к `GET /api/v1/users` и `GET /api/v1/users/:id/progress`.
