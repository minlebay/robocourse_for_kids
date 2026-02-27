# Домен: Администрирование

**Обновлено:** 2026-02-27

## Назначение

Управление пользователями платформы: создание, блокировка, сброс паролей, просмотр статистики и активности. Доступно только пользователям с ролью `administrator`.

**Создание администратора:** роль `administrator` нельзя выбрать в форме «Создать пользователя» в админке (там только student и teacher). Создать администратора можно только через прямой вызов API (`POST /api/v1/admin/users` с `"role": "administrator"`) или через миграцию/скрипт.

## Эндпоинты

Все маршруты имеют префикс `/api/v1/admin` и требуют JWT с ролью `administrator`.

### GET /api/v1/admin/users

Список всех пользователей платформы.

Ответ: массив пользователей. У каждого есть `role` (основная роль для отображения) и `roles` (полный список ролей из `user_roles`).
```json
[
  {
    "id": "uuid",
    "login": "string",
    "name": "string",
    "role": "student|teacher|administrator",
    "roles": ["student"],
    "theme": "default",
    "email": "string|null",
    "must_change_password": false,
    "is_blocked": false,
    "created_at": "2026-02-20T10:00:00Z"
  }
]
```

### POST /api/v1/admin/users

Создание нового пользователя. Автоматически устанавливает `must_change_password=true`.

Тело запроса:
```json
{
  "login": "string (required, 3-50 chars)",
  "password": "string (required, 6-72 chars)",
  "name": "string (required, max 200 chars)",
  "role": "student|teacher|administrator (default: student)",
  "email": "string (optional, validated via net/mail.ParseAddress)"
}
```

Ответ (201):
```json
{
  "user": { ...user fields... },
  "temp_password": "plaintext password to share with user"
}
```

Ошибки:
- 400 — невалидный логин/пароль/имя/роль/email
- 400 — логин уже существует

### DELETE /api/v1/admin/users/:id

Удаление пользователя. Нельзя удалить самого себя (400).

Ответ: 204 No Content.

### POST /api/v1/admin/users/:id/block

Блокировка или разблокировка пользователя. Нельзя заблокировать самого себя (400).

Тело запроса:
```json
{ "block": true }
```

Ответ:
```json
{ "blocked": true }
```

### POST /api/v1/admin/users/:id/reset-password

Генерирует случайный 10-символьный пароль (буквы + цифры) и устанавливает `must_change_password=true`.

Ответ:
```json
{ "temp_password": "aB3xY7kLmN" }
```

### GET /api/v1/admin/stats

Общая статистика платформы.

Ответ:
```json
{
  "users": 42,
  "modules": 8,
  "lessons": 64
}
```

### GET /api/v1/admin/activity

Последние 20 зарегистрировавшихся пользователей (по `created_at DESC`).

Ответ: массив пользователей (те же поля, что и в GET /admin/users).

## Бизнес-логика

### Temp Password Flow

1. Администратор создаёт пользователя через `POST /admin/users` или сбрасывает пароль через `POST /admin/users/:id/reset-password`.
2. Сервер устанавливает `must_change_password=true` в БД.
3. Новый JWT (при следующем логине) содержит `must_change_password=true` в claims.
4. Middleware `RequireFreshPassword` блокирует все запросы с 403 `{ "error": "password_change_required" }`, кроме `POST /api/v1/auth/change-password`.
5. Пользователь отправляет `POST /api/v1/auth/change-password` с текущим и новым паролем.
6. Сервер обновляет пароль, устанавливает `must_change_password=false`, возвращает новый JWT.
7. С новым JWT пользователь получает полный доступ.

### Блокировка пользователей

- `is_blocked=true` проверяется только при `POST /api/v1/auth/login`.
- Заблокированный пользователь получает 403 при попытке входа — существующие JWT до истечения (1 час) продолжают работать.
- Для немедленного отзыва доступа: заблокировать + сбросить пароль через `POST /admin/users/:id/reset-password`.

### Генерация временных паролей

Используется `crypto/rand` для криптографически безопасной генерации. Алфавит: `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`. Длина: 10 символов.

## Миграции

- **000031** — добавляет значение `administrator` в enum `user_role`. Выполняется отдельно от DDL (PostgreSQL не поддерживает `ALTER TYPE ADD VALUE` внутри транзакции вместе с другими DDL командами).
- **000032** — добавляет колонки `email`, `must_change_password`, `is_blocked` в таблицу `users`. Устанавливает роль `administrator` для пользователя с логином `admin`.

## Безопасность

- Только роль `administrator` имеет доступ к маршрутам `/api/v1/admin/*`.
- Администратор не может удалить или заблокировать самого себя.
- Email валидируется через `net/mail.ParseAddress` (строгая валидация RFC 5322).
- Временные пароли хешируются bcrypt перед сохранением в БД.
