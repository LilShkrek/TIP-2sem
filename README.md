# Практическое занятие №2

## Цель

Освоить создание простого микросервиса на Go с использованием gRPC: описать контракт в Protocol Buffers, сгенерировать Go-код, запустить gRPC-сервер и выполнить клиентские вызовы.

## Задание

Реализовать учебный сервис `StudentService` с методами:

- `Ping` — проверка доступности сервера;
- `GetStudentByID` — получение данных студента по идентификатору.

Сервис хранит тестовые данные в памяти и слушает порт `50051`.

## Структура проекта

```text
.
├── proto/
│   └── student.proto
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── client/
│       └── main.go
├── internal/
│   └── student/
│       ├── data.go
│       ├── service.go
│       └── service_test.go
├── gen/
│   └── studentpb/
│       ├── student.pb.go
│       └── student_grpc.pb.go
├── go.mod
├── go.sum
└── README.md
```

## Генерация protobuf-кода

Для генерации нужны `protoc`, `protoc-gen-go` и `protoc-gen-go-grpc`.

Установка Go-плагинов:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Генерация из корня проекта:

```bash
protoc --proto_path=proto \
  --go_out=. --go_opt=module=example.com/pz2-grpc \
  --go-grpc_out=. --go-grpc_opt=module=example.com/pz2-grpc \
  proto/student.proto
```

## Запуск сервера

```bash
go run ./cmd/server
```

Ожидаемый вывод:

```text
gRPC server started on :50051
```

## Запуск клиента

Во втором терминале:

```bash
go run ./cmd/client
```

Пример вывода:

```text
Ping response: Server received: hello grpc
Student: id=1, full_name=Иванов Иван Иванович, group=ИВБО-01-25, email=ivanov@example.com
NotFound check: student with id=999 was not found
```

## Проверка ошибки NotFound

Клиент дополнительно вызывает `GetStudentByID` с `id=999`. Такого студента нет в памяти сервера, поэтому сервер возвращает gRPC-ошибку с кодом `NotFound`.

Проверить это можно запуском клиента:

```bash
go run ./cmd/client
```

## Проверка проекта

```bash
gofmt -w ./cmd ./internal ./gen
go test ./...
go vet ./...
```
