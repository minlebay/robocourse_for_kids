# Глубокое ревью ролевой системы

**Дата:** 2026-02-27

## 1. Обзор архитектуры

**Источник прав:** только таблица `user_roles` (user_id, role). Колонка `users.role` больше не заполняется при создании пользователей (миграция 000041 сделала её nullable); в ответах API поле `role` вычисляется из `user_roles` (PrimaryRole: administrator > teacher > course_owner > student).

- **Таблица `user_roles`** — единственный источник прав; роли: `student`, `teacher`, `course_owner`, `administrator`.
- **Таблица `users.role`** — nullable, deprecated; может быть заполнена только у старых строк (до миграции 000041).

Авторизация и проверки прав опираются на **роли из JWT** (массив `roles` из `user_roles`). В контекст Gin middleware записывает `user_roles` (массив строк). Все проверки делаются по этому массиву и по `modules.owner_id` для редактирования курсов.

---

## 2. Что сделано хорошо

### 2.1 Единый источник прав в запросе

- JWT содержит `roles` из `user_roles`; при каждом запросе роли кладутся в контекст.
- Проверки в middleware (`RequireTeacher`, `RequireAdmin`) и в handler’ах lessons используют один и тот же список ролей из контекста.
- Нет дублирования логики «кто такой учитель»: константы в `users` (RoleTeacher, RoleAdministrator), один хелпер `hasRole` в middleware и в lessons.

### 2.2 Разделение teacher / admin

- **Teacher:** доступ к списку учеников, прогрессу, дашборду; редактирование только своих курсов (или через admin).
- **Admin:** полный доступ, включая `/api/v1/admin/*`, блокировка, сброс пароля.
- `RequireTeacher` пропускает и teacher, и administrator; `RequireAdmin` — только administrator. Цепочка прав согласована.

### 2.3 Владение курсами (ownership)

- Право редактировать модуль/урок проверяется в handler’ах lessons через `canEditModule(roles, mod, userID)`:
  - либо есть роль `administrator`,
  - либо `mod.OwnerID == userID`.
- Роль `course_owner` в БД для этой проверки не используется — достаточно `owner_id`. Это упрощает модель и избегает рассинхрона «роль vs владелец модуля».

### 2.4 Регистрация и инвайт-код

- Роль teacher при регистрации защищена `TEACHER_INVITE_CODE`; при пустом коде регистрация учителей отключена.
- Administrator через публичную регистрацию создать нельзя (только student/teacher); админ создаётся миграцией или через admin API. Это соответствует документации.

### 2.5 Один источник прав при создании пользователя

- `Repo.Create` пишет только в `user_roles` (и в `users` без поля role). Регистрация и AdminCreate используют один путь; дублирование с `users.role` устранено (миграция 000041).

### 2.6 Тесты

- Middleware: отдельно проверены RequireAuth, RequireTeacher (teacher + administrator), RequireAdmin (только administrator), RequireFreshPassword, Auth при ошибке IsUserBlocked.
- Lessons: в тестах выставляется `user_roles` (например, administrator) для проверки canEditModule.
- Фронт: RequireTeacher проверяет teacher/student/неавторизован; `roles.ts` учитывает и `role`, и `roles`.

### 2.7 Фронтенд

- Единые хелперы в `shared/roles.ts`: `hasTeacherAccess`, `hasAdminAccess`, `canEditModule` (по `is_owner`/`owner_id` и admin). Навигация и роуты используют их последовательно.
- Поддержка и одного `role`, и массива `roles` в `hasTeacherAccess`/`hasAdminAccess` совместима с текущим и возможным будущим API.

---

## 3. Замечания и риски

### 3.1 ~~Дублирование: `users.role` и `user_roles`~~ (устранено)

- **Сделано:** миграция 000041 сделала `users.role` nullable; при создании пользователя пишется только `user_roles`. Поле `role` в ответах API вычисляется через `PrimaryRole(roles)` из `user_roles`. Дублирование устранено.

### 3.2 Роль `course_owner` не используется в коде

- В миграции 000040 в `user_roles` есть CHECK с `course_owner`, но ни Repo.Create, ни AdminCreate эту роль не пишут.
- Проверка прав редактирования курса идёт только по `modules.owner_id` и роли `administrator`; по роли `course_owner` ничего не проверяется.
- **Итог:** роль `course_owner` в БД — задел на будущее или для отображения. Документация (auth.md) описывает её как «владелец курса (привязка через modules.owner_id)» — с точки зрения прав это сейчас именно owner_id, а не запись в user_roles.
- **Рекомендация:** либо не добавлять `course_owner` в `user_roles` до появления явной необходимости (например, отображение «владелец курса» в профиле), либо зафиксировать в docs, что права редактирования определяются только `owner_id` и ролью administrator.

### 3.3 ~~Admin API: список пользователей без массива `roles`~~ (исправлено)

- **Сделано:** сервис `ListAll` и `List` (и `GetActivity`) вызывают `attachRoles`: подгружают роли через `GetRolesByUserIDs` и заполняют у каждого пользователя `Roles` и `Role` (primary). GET /api/v1/admin/users и GET /api/v1/users теперь возвращают поле `roles` у каждого пользователя.

### 3.4 Админка: создание администратора через UI невозможно (задокументировано)

- Бэкенд AdminCreate допускает роль `administrator`; фронт предлагает только `student` и `teacher`. В docs/context/auth.md и admin.md явно указано, что администратор создаётся только через миграцию или прямой вызов API.

### 3.5 Типы на фронтенде

- `User.role` в `types.ts`: `'student' | 'teacher' | 'administrator'` — совпадает с тем, что может прийти в API. Тип `UserRole` включает `course_owner` — нормально, если в будущем API начнёт отдавать эту роль в `roles`.
- `AdminCreateUserRequest.role`: только `'student' | 'teacher'` — соответствует текущему UI; при добавлении возможности создавать администратора через админку тип нужно расширить.

### 3.6 ~~Комментарий course_owner/owner_id~~ (добавлен)

- В `lessons/handler.go` у функции `canEditModule` добавлен комментарий: проверка прав основана только на роли administrator и `modules.owner_id`; роль `course_owner` в user_roles не проверяется.

### 3.7 Маршруты модулей/уроков и RequireTeacher

- PUT/DELETE lessons, POST/DELETE modules защищены `RequireTeacher` и затем внутри handler’а — `canEditModule`. То есть сначала «хотя бы учитель», затем «владелец или admin». Цепочка корректна; лишний доступ у «просто учителя» к чужим курсам не выдаётся.

### 3.8 Заблокированные пользователи

- При блокировке проверка выполняется в Auth middleware при каждом запросе (IsUserBlocked). Существующие JWT у заблокированного пользователя остаются валидными до истечения TTL. В docs это описано; при необходимости немедленного отзыва можно доп. использовать сброс пароля (must_change_password).

---

## 4. Сводная таблица проверок прав

| Действие / Эндпоинт              | Middleware / условие      | Доп. проверка в handler              |
|----------------------------------|---------------------------|--------------------------------------|
| GET /admin/*                     | RequireAdmin              | —                                    |
| GET/DELETE /users, GET progress  | RequireTeacher            | ValidateDeleteUser (только студенты) |
| POST/PUT/DELETE modules/lessons  | RequireTeacher            | canEditModule (owner или admin)      |
| POST /modules                    | RequireTeacher            | — (создатель становится owner)       |
| Остальные JWT-эндпоинты          | RequireAuth               | —                                    |

---

## 5. Рекомендации по доработкам (выполнено)

1. **Документация:** в auth.md и admin.md указано, что администратор создаётся только через миграцию или прямой вызов API.
2. **Список ролей в API:** GET /api/v1/admin/users, GET /api/v1/users и GET /api/v1/admin/activity возвращают у каждого пользователя поле `roles` (и вычисляемое `role`).
3. **Комментарий в коде:** в lessons/handler.go у `canEditModule` зафиксировано, что проверка прав — только по `owner_id` и роли administrator.
4. **Один источник прав:** при создании пользователя пишется только `user_roles`; `users.role` не заполняется (миграция 000041).

---

## 6. Заключение

Ролевая система приведена к одному источнику прав (`user_roles`): дублирование с `users.role` устранено (миграция 000041), в ответах API поле `role` вычисляется из `user_roles`. Списки пользователей (admin/users, users, activity) возвращают массив `roles` у каждого пользователя. В документации и в коде зафиксировано: администратор создаётся только через миграцию или API; проверка прав редактирования курса — только по `modules.owner_id` и роли administrator.
