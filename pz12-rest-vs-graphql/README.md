# Практическая работа N12: REST и GraphQL для Task

## Цель

Реализовать один и тот же функционал для сущности `Task` двумя способами: через REST API и через GraphQL API. На одном пользовательском сценарии нужно сравнить количество запросов, состав данных, обработку ошибок, документирование, тестирование и кэширование.

## Единый сценарий сравнения

Экран списка задач использует поля `id`, `title`, `done`. Экран деталей использует поля `id`, `title`, `description`, `done`. Действия изменения данных: создание новой задачи и обновление признака `done`.

Оба API работают с одной моделью `Task`, общим in-memory repository и общим service layer. Поэтому REST и GraphQL видят одинаковые данные в рамках одного запущенного процесса.

## Структура проекта

```text
pz12-rest-vs-graphql/
├── cmd/server/              # запуск общего REST + GraphQL сервера
├── graph/                   # схема и резолверы gqlgen
├── internal/rest/           # REST handlers
├── internal/task/           # модель, repository и service layer
├── gqlgen.yml               # конфигурация gqlgen
├── go.mod
└── README.md
```

## Запуск

```bash
cd pz12-rest-vs-graphql
go run ./cmd/server
```

По умолчанию сервер запускается на порту `8082`.

REST API доступен по адресу:

```text
http://localhost:8082/v1/tasks
```

GraphQL endpoint доступен по адресу:

```text
http://localhost:8082/query
```

GraphQL Playground доступен по адресу:

```text
http://localhost:8082/
```

## REST-примеры

Получить список задач:

```bash
curl -s http://localhost:8082/v1/tasks
```

Получить одну задачу:

```bash
curl -s http://localhost:8082/v1/tasks/t_001
```

Создать задачу:

```bash
curl -s -X POST http://localhost:8082/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Сравнить REST и GraphQL","description":"Практическая работа N12"}'
```

Обновить `done`:

```bash
curl -s -X PATCH http://localhost:8082/v1/tasks/t_001 \
  -H "Content-Type: application/json" \
  -d '{"done":true}'
```

Пример ошибки:

```bash
curl -i http://localhost:8082/v1/tasks/unknown
```

Ожидаемый результат: HTTP `404 Not Found` и тело `{"error":"task not found"}`.

## GraphQL-примеры

Список задач только с полями для экрана списка:

```graphql
query {
  tasks {
    id
    title
    done
  }
}
```

Детали задачи:

```graphql
query GetTask($id: ID!) {
  task(id: $id) {
    id
    title
    description
    done
  }
}
```

Переменные:

```json
{
  "id": "t_001"
}
```

Создание задачи:

```graphql
mutation Create($input: CreateTaskInput!) {
  createTask(input: $input) {
    id
    title
    description
    done
  }
}
```

Переменные:

```json
{
  "input": {
    "title": "Сравнить REST и GraphQL",
    "description": "Практическая работа N12"
  }
}
```

Обновление `done`:

```graphql
mutation Update($id: ID!, $input: UpdateTaskInput!) {
  updateTask(id: $id, input: $input) {
    id
    title
    description
    done
  }
}
```

Переменные:

```json
{
  "id": "t_001",
  "input": {
    "done": true
  }
}
```

Пример ошибки:

```graphql
query {
  task(id: "unknown") {
    id
    title
  }
}
```

Ожидаемый результат: HTTP-ответ обычно остаётся `200 OK`, а ошибка находится в поле `errors` JSON-ответа.

## Сравнение

| Критерий | REST | GraphQL |
|---|---|---|
| Количество запросов | Для сценария список -> детали -> обновление нужно 3 запроса: `GET`, `GET`, `PATCH`. | Для того же сценария нужно 3 операции: query списка, query деталей, mutation обновления. |
| Объём данных | `GET /v1/tasks` возвращает всю модель, включая `description`, хотя экран списка использует только `id`, `title`, `done`. | Клиент сам запрашивает только `id`, `title`, `done` для списка, поэтому лишних полей нет. |
| Обработка ошибок | Ошибка выражается HTTP-статусом и JSON-телом, например `404` и `{"error":"task not found"}`. | Ошибка приходит в поле `errors`; HTTP-статус часто остаётся `200`, поэтому клиент должен анализировать JSON-ответ. |
| Документирование | Endpoint и HTTP-методы легко описывать через таблицу маршрутов или OpenAPI. | Схема GraphQL сама показывает типы, query и mutation, но клиенту нужно понимать язык запросов. |
| Тестирование | REST удобно проверять через `curl`, `httptest` и статусы HTTP. | GraphQL удобно тестировать через отправку query/mutation и проверку полей `data` и `errors`. |
| Кэширование | Концептуально проще: разные URL и методы хорошо ложатся на стандартный HTTP-кэш. | Сложнее: endpoint один, поэтому чаще нужны persisted queries, client-side cache или кэширование на уровне данных. |
| Простота учебной реализации | Выше, потому что достаточно маршрутов, JSON и HTTP-статусов. | Ниже, потому что нужны схема, генерация `gqlgen` и резолверы. |
| Гибкость клиента | Ниже, потому что форму ответа в основном задаёт сервер. | Выше, потому что клиент выбирает поля под конкретный экран. |

## Итоговый вывод

В выбранном сценарии REST и GraphQL требуют одинаковое количество сетевых обращений: список, детали и изменение данных. Главное отличие проявляется не в числе запросов, а в точности возвращаемых данных. REST endpoint списка возвращает всю модель задачи, включая `description`, хотя экран списка это поле не использует. GraphQL позволяет запросить только `id`, `title`, `done`, поэтому лучше подходит для экранов с разным набором полей. При этом REST проще читать, тестировать через `curl` и анализировать по HTTP-статусам. GraphQL требует больше инфраструктуры: схемы, генерации кода и резолверов. Обработка ошибок в REST в учебном CRUD-сервисе выглядит понятнее, потому что статус `404` сразу отражает проблему. В GraphQL ошибка находится в поле `errors`, поэтому клиентская логика должна быть внимательнее. Для простой системы задач REST выглядит более практичным решением. GraphQL становится полезнее, когда у клиента много разных экранов и важно точно выбирать структуру ответа.

## Проверка

```bash
gofmt -w .
go run github.com/99designs/gqlgen generate
go test ./...
go vet ./...
```
