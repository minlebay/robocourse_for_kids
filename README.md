# Learn Kids — персональная образовательная платформа для обучения робототехнике

Проект полностью навайбкожен и является моим экспериментом :)

Персональная образовательная платформа для обучения детей и взрослых робототехнике: от основ (мигающий светодиод) до сложных проектов (робот-паук, квадрокоптер) на Arduino / Raspberry Pi. Доступ в домашней локальной сети.

## Требования

- **Docker** и **docker compose** — для запуска всего стека.
- Опционально для локальной разработки: **Go 1.21+**, **Node 20+**.

## Быстрый старт

1. Клонируйте репозиторий и перейдите в каталог проекта.
2. Скопируйте пример окружения и при необходимости отредактируйте переменные:
   ```bash
   cp .env.example .env
   ```
3. Запустите сервисы:
   ```bash
   docker compose up --build
   ```
4. Откройте в браузере: **http://localhost:8888** — доступ через Nginx (reverse proxy).

API: **http://localhost:8888/api/v1**. Спецификация OpenAPI: [backend/api/openapi.yaml](backend/api/openapi.yaml). Интерактивная документация: `GET http://localhost:8888/api/docs`.

## Сервисы

- **nginx** — reverse proxy, порт 8888. Принимает запросы только из подсети 192.168.1.0/24 и localhost.
- **api** — бэкенд на Go (Gin): REST API, раздача статики фронта, миграции БД при старте. Порт 8080 только внутри Docker-сети.
- **db** — PostgreSQL 15. Данные хранятся в volume `db_data`.

## Разработка без Docker

### Бэкенд

```bash
cd backend
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/learn_kids?sslmode=disable"
go run ./cmd/api
```

БД должна быть запущена (например, `docker compose up -d db`).

### Фронтенд

В Docker фронт отдаётся с **http://localhost:8888** (через Nginx). Для разработки UI можно запустить Vite отдельно:

```bash
cd frontend
npm install
npm run dev
```

Тогда фронт будет на отдельном порту с проксированием `/api` на бэкенд (`vite.config.ts`). Для работы «всё из контейнера» используйте **http://localhost:8888**.

## Перенос на другой хост

Скопируйте на новый хост:

- Исходный репозиторий (или образы и `docker-compose.yml`).
- Файл `.env` с актуальными `DATABASE_URL`, `JWT_SECRET`, при необходимости `FRONTEND_ORIGIN`.

Запустите `docker compose up -d`. Volume `db_data` сохраняет данные БД — при переносе его можно перенести или оставить на старом хосте и заново поднять БД.

Подробнее: [docs/deployment.md](docs/deployment.md).

## Тесты и .test.env

В корне проекта лежит файл **.test.env** с тестовыми переменными окружения (`DATABASE_URL`, `JWT_SECRET`, `PORT`, `FRONTEND_ORIGIN`, `GEMINI_API_KEY`). Интеграционные тесты бэкенда при запуске автоматически подхватывают его (из `../.test.env` при запуске из `backend/`). Сервис **api** в `docker compose` тоже загружает `.test.env` (в дополнение к `.env`), поэтому ключ Gemini для чата с ИИ подхватится из `.test.env`.

- Запуск тестов бэкенда (сначала поднимите БД: `docker compose up -d db`):
  ```bash
  cd backend && go test ./...
  ```
- Запуск тестов фронтенда: `cd frontend && npm run test`

Значения в `.test.env` подставлены так, чтобы тесты и локальный запуск API работали с БД на localhost:5432.

## Контекст для разработчиков и агентов

В каталоге [docs/context/](docs/context/) хранятся текстовые описания по доменам (каталог уроков, авторизация, прогресс, дашборд родителя, контракт API). Перед изменением кода домена рекомендуется читать соответствующий файл; после изменения API или форматов — обновлять описание. Правила использования: [docs/context/README.md](docs/context/README.md).

## Документация API

- Спецификация OpenAPI 3.0: [backend/api/openapi.yaml](backend/api/openapi.yaml).
- Краткая сводка эндпоинтов: [docs/context/api-contract.md](docs/context/api-contract.md).
- Как смотреть и использовать API: [docs/openapi.md](docs/openapi.md).
