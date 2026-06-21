# Практическое занятие №13

## Тема и цель

Тема: подключение к RabbitMQ, отправка и получение сообщений.

Цель: реализовать учебный сценарий, в котором сервис `tasks` создаёт задачу через HTTP и публикует событие `task.created` в RabbitMQ, а отдельный `worker` получает сообщение, логирует его и подтверждает обработку.

## RabbitMQ, producer и consumer

RabbitMQ используется как брокер сообщений между HTTP-сервисом и фоновым обработчиком.

- `tasks` — producer: после успешного создания задачи публикует событие в очередь `task_events`.
- `worker` — consumer: читает сообщения из очереди, десериализует JSON, логирует данные события и отправляет подтверждение обработки.

## Структура проекта

```text
pz13-rabbitmq/
├── deploy/
│   └── rabbit/
│       └── docker-compose.yml
├── internal/
│   ├── env/
│   └── events/
├── services/
│   ├── tasks/
│   │   ├── cmd/tasks/main.go
│   │   └── internal/
│   │       ├── amqp/
│   │       ├── http/
│   │       └── service/
│   └── worker/
│       ├── cmd/worker/main.go
│       └── internal/consumer/
├── go.mod
├── go.sum
└── README.md
```

## Формат события

```json
{
  "event": "task.created",
  "task_id": "t_001",
  "ts": "2026-03-26T10:20:30Z",
  "request_id": "pz13-001",
  "producer": "tasks",
  "version": "1.0"
}
```

`request_id` заполняется из заголовка `X-Request-ID`, если он передан.

## Надёжность обработки

Очередь `task_events` объявляется как durable, поэтому её описание сохраняется при перезапуске RabbitMQ.

Сообщения публикуются с `DeliveryMode: amqp.Persistent`, чтобы RabbitMQ сохранял их надёжнее, чем transient-сообщения.

Worker использует ручной `ack`: после успешной обработки вызывается `Ack(false)`. Если JSON некорректный, вызывается `Nack(false, false)`, и сообщение не возвращается в очередь.

Для consumer установлен `prefetch = 1`, чтобы worker получал не больше одного неподтверждённого сообщения одновременно.

## Переменные окружения

| Переменная | Значение по умолчанию | Назначение |
| --- | --- | --- |
| `RABBIT_URL` | `amqp://guest:guest@localhost:5672/` | адрес RabbitMQ |
| `QUEUE_NAME` | `task_events` | имя очереди |
| `TASKS_PORT` | `8082` | порт HTTP-сервиса `tasks` |

## Запуск RabbitMQ

```bash
cd deploy/rabbit
docker compose up -d
docker compose ps
```

RabbitMQ Management UI:

```text
http://localhost:15672
```

Логин и пароль:

```text
guest / guest
```

## Запуск worker

```bash
go run ./services/worker/cmd/worker
```

## Запуск сервиса tasks

```bash
TASKS_PORT=8082 go run ./services/tasks/cmd/tasks
```

## Пример запроса

```bash
curl -i -X POST http://localhost:8082/v1/tasks \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: pz13-001" \
  -d '{"title":"Rabbit","description":"publish event"}'
```

## Ожидаемый ответ сервиса

```json
{
  "id": "t_001",
  "title": "Rabbit",
  "description": "publish event",
  "created_at": "2026-03-26T10:20:30Z"
}
```

HTTP-статус: `201 Created`.

## Ожидаемый лог worker

```text
received event=task.created task_id=t_001 ts=2026-03-26T10:20:30Z request_id=pz13-001
```

## Проверка через Management UI

В RabbitMQ Management UI нужно открыть раздел `Queues and Streams` и убедиться, что существует очередь `task_events`. В карточке очереди видны количество сообщений, активные consumers, ready-сообщения и unacked-сообщения.

## Режим best effort

Публикация события работает в режиме `best effort`: задача создаётся независимо от результата отправки сообщения. Если RabbitMQ недоступен или публикация завершилась ошибкой, сервис пишет ошибку в лог, но клиент всё равно получает успешный ответ о создании задачи.

## Команды тестирования

```bash
gofmt -w .
go mod tidy
go test ./...
go vet ./...
```
