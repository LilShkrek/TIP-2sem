# Практическое занятие №14

## Тема и цель работы

Тема: реализация очереди задач по модели producer-consumer с использованием RabbitMQ.

Цель работы: освоить постановку задач в очередь, обработку задач отдельным worker-процессом, ограниченные повторные попытки, отправку проблемных сообщений в DLQ и базовую идемпотентность по `message_id`.

## Отличие задачи от события

Событие сообщает, что в системе уже что-то произошло. Задача описывает работу, которую нужно выполнить. В этой работе сообщение в очереди является задачей: оно может обрабатываться дольше HTTP-запроса, завершаться ошибкой и требовать повторной попытки.

## Producer и consumer

Producer: HTTP-сервис `tasks`. Он принимает запрос `POST /v1/jobs/process-task`, формирует job и публикует ее в очередь `task_jobs`.

Consumer: отдельный процесс `worker`. Он читает сообщения из `task_jobs`, выполняет имитацию тяжелой операции, подтверждает успешные сообщения через `ack`, а при ошибках публикует задачу заново или отправляет ее в `task_jobs_dlq`.

## Структура проекта

```text
pz14-task-queue/
  deploy/
    rabbit/
      docker-compose.yml
  internal/
    env/
    jobs/
  services/
    tasks/
      cmd/tasks/main.go
      internal/
        amqp/
        http/
        jobs/
    worker/
      cmd/worker/main.go
      internal/
        consumer/
        store/
  go.mod
  README.md
```

## Формат сообщения

```json
{
  "job": "process_task",
  "task_id": "t_001",
  "attempt": 1,
  "message_id": "uuid-here"
}
```

- `job` - тип задачи, для этой работы используется `process_task`;
- `task_id` - идентификатор задачи;
- `attempt` - номер текущей попытки обработки;
- `message_id` - уникальный идентификатор сообщения.

## Attempt и message_id

При первичной публикации `attempt` всегда равен `1`, а `message_id` генерируется как UUID через `crypto/rand`.

Worker проверяет `message_id`. Если он пустой, сообщение считается некорректным и отправляется в DLQ. Если `message_id` уже есть в in-memory хранилище успешно обработанных сообщений, worker считает сообщение дубликатом, не выполняет работу повторно и сразу отправляет `ack`.

## Retries

Максимальное число попыток обработки: `3`.

Если обработка завершилась ошибкой, worker увеличивает `attempt`. Пока новое значение `attempt` не больше `MAX_ATTEMPTS`, worker публикует сообщение заново в `task_jobs` и только после успешной публикации подтверждает исходное сообщение через `ack`.

Если повторная публикация не удалась, worker не подтверждает исходное сообщение и делает `nack` с requeue, чтобы задача не была потеряна.

## DLQ

DLQ-очередь: `task_jobs_dlq`.

Если после ошибки новое значение `attempt` превышает лимит, worker публикует сообщение в `task_jobs_dlq` и только после успешной публикации подтверждает исходное сообщение через `ack`.

Если публикация в DLQ не удалась, worker делает `nack` с requeue. Некорректный JSON worker отклоняет через `nack` без повторной постановки в очередь. Это поведение выбрано для учебной простоты: такое сообщение нельзя надежно преобразовать в валидный payload для DLQ.

## Идемпотентность

Worker хранит успешно обработанные `message_id` в памяти. Хранилище защищено `sync.RWMutex`, поэтому безопасно для конкурентного доступа. Повторное сообщение с уже обработанным `message_id` подтверждается через `ack` без повторного выполнения работы.

## Переменные окружения

| Переменная | Значение по умолчанию | Назначение |
| --- | --- | --- |
| `RABBIT_URL` | `amqp://guest:guest@localhost:5672/` | адрес RabbitMQ |
| `TASK_QUEUE` | `task_jobs` | основная очередь задач |
| `DLQ_QUEUE` | `task_jobs_dlq` | очередь проблемных сообщений |
| `TASKS_PORT` | `8082` | порт HTTP-сервиса `tasks` |
| `MAX_ATTEMPTS` | `3` | максимальное число попыток обработки |

## Запуск RabbitMQ

```bash
cd pz14-task-queue/deploy/rabbit
docker compose up -d
docker compose ps
```

Management UI:

```text
http://localhost:15672
```

Логин и пароль:

```text
guest / guest
```

## Запуск worker

```bash
cd pz14-task-queue
go run ./services/worker/cmd/worker
```

## Запуск tasks

```bash
cd pz14-task-queue
TASKS_PORT=8082 go run ./services/tasks/cmd/tasks
```

## Пример успешной задачи

```bash
curl -i -X POST http://localhost:8082/v1/jobs/process-task \
  -H "Content-Type: application/json" \
  -d '{"task_id":"t_001"}'
```

Ожидаемый ответ:

```json
{
  "status": "accepted",
  "task_id": "t_001",
  "message_id": "..."
}
```

Ожидаемый лог worker:

```text
worker ... processed successfully message_id=... task_id=t_001 attempt=1
```

## Пример задачи с ошибкой

```bash
curl -i -X POST http://localhost:8082/v1/jobs/process-task \
  -H "Content-Type: application/json" \
  -d '{"task_id":"t_fail"}'
```

Ожидаемое поведение:

```text
attempt=1 -> ошибка, повторная публикация attempt=2
attempt=2 -> ошибка, повторная публикация attempt=3
attempt=3 -> ошибка, отправка в task_jobs_dlq
```

Ожидаемые логи worker:

```text
worker ... processing failed message_id=... task_id=t_fail attempt=1 error=simulated processing error
worker ... republished retry message_id=... task_id=t_fail next_attempt=2
worker ... processing failed message_id=... task_id=t_fail attempt=2 error=simulated processing error
worker ... republished retry message_id=... task_id=t_fail next_attempt=3
worker ... processing failed message_id=... task_id=t_fail attempt=3 error=simulated processing error
worker ... sent to DLQ message_id=... task_id=t_fail exceeded_attempt=4
```

## Проверка очередей через Management UI

1. Откройте `http://localhost:15672`.
2. Перейдите в раздел `Queues and Streams`.
3. Проверьте наличие очередей `task_jobs` и `task_jobs_dlq`.
4. После запроса с `task_id=t_fail` проверьте, что в `task_jobs_dlq` появилось сообщение.

## Команды проверки

```bash
cd pz14-task-queue
gofmt -w .
go mod tidy
go test ./...
go vet ./...
```
