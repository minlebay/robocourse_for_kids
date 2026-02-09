# Документация API (OpenAPI)

## Где лежит спецификация

Файл спецификации OpenAPI 3.0: **[backend/api/openapi.yaml](../backend/api/openapi.yaml)** (от корня репозитория).

В нём описаны все эндпоинты, тела запросов/ответов и коды ответов.

## Как смотреть

1. **В редакторе** — откройте `backend/api/openapi.yaml` в VS Code с расширением OpenAPI (Swagger) или в любом редакторе с подсветкой YAML.
2. **Через сервер** — при запущенном API запрос `GET http://localhost:8080/api/docs` отдаёт содержимое файла openapi.yaml (если путь настроен в сервере).
3. **Swagger UI** — скопируйте содержимое openapi.yaml на [editor.swagger.io](https://editor.swagger.io) или запустите локально Swagger UI с указанием пути к файлу.

## Примеры запросов (curl)

- Проверка здоровья:
  ```bash
  curl http://localhost:8080/api/v1/health
  ```
- Регистрация:
  ```bash
  curl -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{"login":"user1","password":"pass","name":"Ученик","role":"student"}'
  ```
- Вход и сохранение токена:
  ```bash
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"login":"user1","password":"pass"}'
  ```
- Список модулей:
  ```bash
  curl http://localhost:8080/api/v1/modules
  ```
- Запрос с авторизацией (подставьте свой токен):
  ```bash
  curl http://localhost:8080/api/v1/auth/me \
    -H "Authorization: Bearer <token>"
  ```

Краткая сводка эндпоинтов и форматов: [docs/context/api-contract.md](context/api-contract.md).
