# Домен: Авторизация (пользователи, роли, JWT)

**Обновлено:** 2026-02-14

## Назначение

Регистрация, вход, выдача JWT; роли `student` и `teacher`. Только учитель имеет доступ к списку пользователей и к прогрессу других учеников. Цветовая схема (тема) сохраняется в профиле пользователя и восстанавливается на любом устройстве.

## Контракт API

- `POST /api/v1/auth/register` — тело: `{ "login", "password", "name", "role?", "invite_code?" }`. Ответ: `{ "user", "token" }`. Rate limit: 10/мин на IP.
- `POST /api/v1/auth/login` — тело: `{ "login", "password" }`. Ответ: `{ "user", "token" }`. Rate limit: 10/мин на IP.
- `GET /api/v1/auth/me` — текущий пользователь (требуется заголовок `Authorization: Bearer <token>`).
- `PATCH /api/v1/auth/me` — обновить профиль. Тело: `{ "theme": "default" | "light" | "cyberpunk" }`. Ответ: обновлённый `user`.

## Валидация

- **Login:** 3–50 символов.
- **Пароль:** 6–72 символа (bcrypt молча обрезает > 72 байт).
- **Роль teacher:** требуется `invite_code` (env `TEACHER_INVITE_CODE`). Если переменная пустая — регистрация учителей отключена.

## Безопасность

- JWT в заголовке `Authorization: Bearer <token>`. Время жизни — 1 час. Claims: `user_id`, `role`, `iss=learn_kids`, `sub=<user_id>`.
- Если заголовок `Authorization` присутствует, но токен невалиден — возвращается 401 (ранее пропускалось молча).
- Роль из JWT сохраняется в `gin.Context` (`user_role`), проверка RequireTeacher без DB-запроса.
- Пароли хранятся как bcrypt-хеш (DefaultCost).
- Timing attack protection при login: dummy bcrypt при отсутствии пользователя.
- Rate limiting на `/auth/login` и `/auth/register`: 10 запросов в минуту на IP.

## Фронтенд

- Форма регистрации показывает поле `invite_code` только при выборе роли `teacher`.
- Навигация в хедере через React Router `<Link>` (без полной перезагрузки страницы).
- Ошибка 429 (rate limit) показывается пользователю как понятное сообщение на русском.
- При 401 (истёкший/невалидный токен) фронт сохраняет текущий URL в sessionStorage и редиректит на `/login`; после успешного входа пользователь возвращается на сохранённую страницу, если путь безопасный (/, /modules/:id, /lessons/:id, /progress, /dashboard).

## Решения

- Роль по умолчанию при регистрации — `student`. Роль `teacher` даёт доступ к `GET /api/v1/users`, `DELETE /api/v1/users/:id` и `GET /api/v1/users/:id/progress`.
- Поле `theme` в таблице `users` — цветовая схема портала. Значения: `default`, `light`, `cyberpunk`.
