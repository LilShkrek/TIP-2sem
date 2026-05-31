# Практическое занятие №1

## Разделение монолита на 2 микросервиса. Взаимодействие через HTTP

Цель работы: освоить базовый подход к разделению backend-приложения на два самостоятельных Go-сервиса и организовать обмен JSON-данными между ними по HTTP.

## Архитектура

Проект состоит из двух независимых сервисов:

- `user-service` хранит пользователей в памяти и отдаёт их по HTTP.
- `order-service` хранит заказы в памяти, отдаёт обычный заказ и агрегированный ответ `заказ + пользователь`.

`order-service` при запросе `/orders/{id}/full` обращается к `user-service` по адресу `http://localhost:8081/users/{user_id}`.

## Структура проекта

```text
.
├── README.md
├── assignments/
│   └── prac1.md
├── user-service/
│   ├── go.mod
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   └── internal/
│       └── user/
│           ├── handler.go
│           ├── handler_test.go
│           ├── model.go
│           └── repo.go
└── order-service/
    ├── go.mod
    ├── cmd/
    │   └── server/
    │       └── main.go
    └── internal/
        └── order/
            ├── client.go
            ├── client_test.go
            ├── handler.go
            ├── handler_test.go
            ├── model.go
            └── repo.go
```

## user-service

Порт: `8081`

Endpoint:

```http
GET /users/{id}
```

Пример ответа:

```json
{"id":1,"name":"Иван Иванов","email":"ivan@example.com"}
```

Ошибки:

- `400 Bad Request` при некорректном `id`;
- `404 Not Found` если пользователь не найден;
- `405 Method Not Allowed` если HTTP-метод не поддерживается.

## order-service

Порт: `8082`

Endpoints:

```http
GET /orders/{id}
GET /orders/{id}/full
```

Пример обычного заказа:

```json
{"id":101,"user_id":1,"item":"Ноутбук","price":79990}
```

Пример агрегированного ответа:

```json
{
  "order": {
    "id": 101,
    "user_id": 1,
    "item": "Ноутбук",
    "price": 79990
  },
  "user": {
    "id": 1,
    "name": "Иван Иванов",
    "email": "ivan@example.com"
  }
}
```

Если `user-service` недоступен при запросе `/orders/{id}/full`, `order-service` возвращает `502 Bad Gateway` и не завершает работу аварийно.

## Запуск

В первом терминале:

```bash
cd user-service
go run ./cmd/server
```

Ожидаемый лог:

```text
user-service started on :8081
```

Во втором терминале:

```bash
cd order-service
go run ./cmd/server
```

Ожидаемый лог:

```text
order-service started on :8082
```

## Проверка через curl

Пользователь:

```bash
curl http://localhost:8081/users/1
```

Заказ:

```bash
curl http://localhost:8082/orders/101
```

Агрегированный ответ:

```bash
curl http://localhost:8082/orders/101/full
```

Ошибка внешнего сервиса:

```bash
curl -i http://localhost:8082/orders/101/full
```

Для этой проверки нужно остановить `user-service`. Ожидается статус `502 Bad Gateway`.

## Проверочные команды

Для `user-service`:

```bash
cd user-service
gofmt -w .
go test ./...
go vet ./...
```

Для `order-service`:

```bash
cd order-service
gofmt -w .
go test ./...
go vet ./...
```

## Вывод

В работе реализовано разделение учебного приложения на два независимых Go-сервиса. `order-service` агрегирует собственные данные с данными из `user-service` через HTTP, а оба сервиса обрабатывают успешные и ошибочные сценарии.
