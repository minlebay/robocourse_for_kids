# Домен: Авторизация (пользователи, роли, JWT)

**Обновлено:** 2026-02-27

## Назначение

Регистрация, вход, выдача JWT; многоролевая система: роли хранятся в таблице `user_roles` (один пользователь может иметь несколько ролей). Роли: `student`, `teacher`, `course_owner`, `administrator`. Только учитель и администратор имеют доступ к списку пользователей и к прогрессу других учеников. Редактирование курсов (модулей и уроков) разрешено только владельцу курса или администратору. Администратор имеет доступ к управлению пользователями (`/api/v1/admin/*`). Цветовая схема (тема) сохраняется в профиле пользователя и восстанавливается на любом устройстве.

## Поля пользователя

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `login` | string | Уникальный логин |
| `name` | string | Отображаемое имя |
| `role` | string | Основная роль для обратной совместимости API; вычисляется из `user_roles` (приоритет: administrator > teacher > course_owner > student). |
| `roles` | string[] | Список ролей из `user_roles` (единственный источник прав). |
| `theme` | string | Цветовая схема (опционально) |
| `email` | string\|null | Email адрес (опционально) |
| `must_change_password` | bool | Требуется смена пароля при следующем входе |
| `is_blocked` | bool | Пользователь заблокирован |
| `created_at` | timestamp | Дата регистрации |

## Контракт API

- `POST /api/v1/auth/register` — тело: `{ "login", "password", "name", "role?", "invite_code?" }`. Ответ: `{ "user", "token" }`. Rate limit: 10/мин на IP.
- `POST /api/v1/auth/login` — тело: `{ "login", "password" }`. Ответ: `{ "user", "token" }`. Rate limit: 10/мин на IP. Если пользователь заблокирован — 403.
- `GET /api/v1/auth/me` — текущий пользователь (требуется заголовок `Authorization: Bearer <token>`).
- `PATCH /api/v1/auth/me` — обновить профиль. Тело: `{ "theme": "..." }`. Ответ: обновлённый `user`. Допустимые темы: `default`, `light`, `cyberpunk`, `contrast-light`, `contrast-dark`, `cream`, `snow`, `midnight`, `forest`.
- `POST /api/v1/auth/change-password` — смена пароля. Тело: `{ "current_password": "...", "new_password": "..." }`. Ответ: `{ "token": "..." }` с новым JWT, где `must_change_password=false`. Требует JWT.

## Валидация

- **Login:** 3–50 символов.
- **Пароль:** 6–72 символа (bcrypt молча обрезает > 72 байт).
- **Имя:** до 200 символов.
- **Роль teacher:** требуется `invite_code` (env `TEACHER_INVITE_CODE`). Если переменная пустая — регистрация учителей отключена.
- **Email:** валидируется через `net/mail.ParseAddress` (не regex).

## Безопасность

- JWT в заголовке `Authorization: Bearer <token>`. Время жизни — 1 час. Claims: `user_id`, `roles` (массив строк), `must_change_password`, `iss=learn_kids`, `sub=<user_id>`.
- **Sliding session:** если до истечения токена осталось меньше 30 минут, при любом успешном запросе с этим токеном сервер добавляет в ответ заголовок `X-New-Token` с новым JWT. Клиент при получении заголовка сохраняет новый токен в localStorage и использует его в следующих запросах — пользователь не перелогинивается при активном использовании.
- Если заголовок `Authorization` присутствует, но токен невалиден — возвращается 401.
- Роли из JWT сохраняются в `gin.Context` (`user_roles` — массив строк), проверка прав по наличию роли в списке.
- Пароли хранятся как bcrypt-хеш (DefaultCost).
- Timing attack protection при login: dummy bcrypt при отсутствии пользователя.
- Rate limiting на `/auth/login` и `/auth/register`: 10 запросов в минуту на IP.

## Логика блокировки (is_blocked)

- Поле `is_blocked` проверяется при каждом вызове `POST /api/v1/auth/login`.
- Если пользователь заблокирован — возвращается 403 `{ "error": "user is blocked" }` без проверки пароля.
- Существующие JWT заблокированных пользователей продолжают работать до истечения (1 час) — сервер не хранит blacklist токенов.
- Чтобы немедленно отозвать доступ, используйте блокировку + принудительную смену пароля.

## Логика must_change_password

- Поле `must_change_password` хранится в JWT claims.
- Middleware `RequireFreshPassword` проверяет флаг при каждом запросе (кроме `POST /api/v1/auth/change-password`).
- При `must_change_password=true` все защищённые эндпоинты возвращают 403 `{ "error": "password_change_required" }`.
- После успешной смены пароля `POST /auth/change-password` возвращает новый JWT с `must_change_password=false`.
- Устанавливается автоматически при создании пользователя через admin API или при сбросе пароля.

## Роли

**Источник прав:** только таблица `user_roles` (user_id, role). Колонка `users.role` больше не заполняется при создании пользователей (nullable, оставлена для совместимости); в ответах API поле `role` вычисляется из `user_roles`.

| Роль | Описание |
|------|----------|
| `student` | Базовая роль. Чтение уроков, прогресс, чат. |
| `teacher` | Просмотр списка пользователей и прогресса учеников. Редактирование контента — только для курсов, где пользователь владелец (или administrator). |
| `course_owner` | Роль в `user_roles` зарезервирована; **права редактирования курса определяются только по `modules.owner_id`** (и по роли administrator). То есть «владелец курса» — это тот, у кого `modules.owner_id = user_id`, а не наличие роли `course_owner` в user_roles. |
| `administrator` | Полный доступ: редактирование любых курсов, управление пользователями через `/api/v1/admin/*`. |

- `RequireTeacher` пропускает пользователей, у которых в `user_roles` есть `teacher` или `administrator`.
- `RequireAdmin` пропускает только при наличии роли `administrator`.
- Редактирование модуля/урока проверяется на бэкенде: разрешено, если пользователь — administrator или владелец данного модуля (`modules.owner_id = user_id`).

## Фронтенд

- Форма регистрации показывает поле `invite_code` только при выборе роли `teacher`.
- Навигация в хедере через React Router `<Link>` (без полной перезагрузки страницы).
- Ошибка 429 (rate limit) показывается пользователю как понятное сообщение на русском.
- При 401 (истёкший/невалидный токен) фронт сохраняет текущий URL в sessionStorage и редиректит на `/login`; после успешного входа пользователь возвращается на сохранённую страницу, если путь безопасный (/, /modules/:id, /lessons/:id, /progress, /dashboard).

## Решения

- Роль по умолчанию при регистрации — `student`. Роль `teacher` даёт доступ к `GET /api/v1/users`, `DELETE /api/v1/users/:id` и `GET /api/v1/users/:id/progress`.
- Роль `administrator` создаётся **только через миграцию или прямой вызов API** `POST /api/v1/admin/users` с `role: "administrator"`. Публичная регистрация и форма «Создать пользователя» в админке не позволяют выбрать роль administrator (в UI доступны только student и teacher).
- Поле `theme` в таблице `users` — цветовая схема портала. Значения: `default`, `light`, `cyberpunk`, `contrast-light`, `contrast-dark`, `cream`, `snow`, `midnight`, `forest`.
- Поле `email` — nullable, используется для уведомлений (не для аутентификации).
