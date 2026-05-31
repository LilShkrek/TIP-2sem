# Практическое занятие №8

## Цель

Освоить базовую настройку CI/CD для backend-сервиса на Go: автоматическую проверку, сборку приложения, сборку Docker-образа и подготовку к возможной доставке в registry.

## Задание

В рамках работы реализован учебный сервис `tasks` и настроен pipeline GitHub Actions. Pipeline запускает проверки Go-кода, собирает приложение и Docker-образ. Публикация образа в registry показана как опциональный пример через secrets и не включена в активное выполнение.

## Структура проекта

```text
.
├── .github/
│   └── workflows/
│       └── ci.yml
├── assignments/
│   └── prac8.md
├── services/
│   └── tasks/
│       ├── cmd/
│       │   └── tasks/
│       │       └── main.go
│       ├── internal/
│       │   └── httpapi/
│       │       ├── router.go
│       │       └── router_test.go
│       ├── .dockerignore
│       ├── Dockerfile
│       └── go.mod
└── README.md
```

## Сервис tasks

Сервис предоставляет минимальный HTTP API:

```text
GET /health
```

Успешный ответ:

```json
{"status":"ok"}
```

По умолчанию сервис запускается на порту `8080`. Порт можно изменить через переменную окружения `PORT`.

## Локальный запуск

```bash
cd services/tasks
go run ./cmd/tasks
```

Проверка:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

## Локальная проверка

```bash
cd services/tasks
gofmt -w ./cmd/tasks ./internal/httpapi
go mod tidy
go test ./...
go vet ./...
go build ./...
docker build -t techip-tasks:local .
```

## Docker

Для сервиса добавлен `Dockerfile` с multi-stage build:

- на первом этапе используется образ `golang:1.23-alpine`, загружаются зависимости и собирается бинарный файл;
- на втором этапе используется минимальный runtime-образ `alpine:3.20`;
- итоговый контейнер запускает бинарный файл `/app/tasks`;
- порт сервиса внутри контейнера: `8080`.

Запуск контейнера:

```bash
cd services/tasks
docker build -t techip-tasks:local .
docker run --rm -p 8080:8080 techip-tasks:local
```

## CI/CD

Для CI/CD выбрана платформа GitHub Actions. Workflow находится в файле `.github/workflows/ci.yml`.

CI отвечает за автоматическую проверку изменений: установку окружения, проверку зависимостей, запуск тестов и сборку приложения.

CD отвечает за подготовку приложения к доставке: упаковку в Docker-образ, возможную публикацию в registry и дальнейший деплой.

Pipeline запускается при `push` и `pull_request` в ветки:

- `prac8`;
- `main`;
- `master`.

Pipeline состоит из двух jobs:

- `test-and-build` проверяет Go-сервис;
- `docker-build` собирает Docker-образ после успешного выполнения `test-and-build`.

## Полный YAML pipeline

```yaml
name: CI Pipeline

on:
  push:
    branches: [ "prac8", "main", "master" ]
  pull_request:
    branches: [ "prac8", "main", "master" ]

jobs:
  test-and-build:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Show Go version
        run: go version

      - name: Download dependencies
        run: go mod tidy
        working-directory: ./services/tasks

      - name: Run tests
        run: go test ./...
        working-directory: ./services/tasks

      - name: Build application
        run: go build ./...
        working-directory: ./services/tasks

  docker-build:
    runs-on: ubuntu-latest
    needs: test-and-build

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build Docker image
        run: docker build -t techip-tasks:${{ github.sha }} .
        working-directory: ./services/tasks
```

## Объяснение jobs

`test-and-build` выполняет основную CI-проверку:

- `Checkout repository` получает код из репозитория;
- `Setup Go` устанавливает Go версии `1.23`;
- `Show Go version` выводит текущую версию Go в лог;
- `Download dependencies` выполняет `go mod tidy`;
- `Run tests` запускает `go test ./...`;
- `Build application` запускает `go build ./...`.

`docker-build` отвечает за упаковку приложения:

- запускается только после успешного `test-and-build`;
- получает код из репозитория;
- настраивает Docker Buildx;
- собирает Docker-образ командой `docker build`.

## working-directory

Проект имеет структуру с сервисом внутри `services/tasks`, поэтому команды Go и Docker выполняются с параметром:

```yaml
working-directory: ./services/tasks
```

Без этого pipeline запускал бы команды из корня репозитория и не нашёл бы `go.mod` и `Dockerfile` сервиса.

## Docker tag

В CI образ собирается с тегом:

```text
techip-tasks:${{ github.sha }}
```

`${{ github.sha }}` — hash коммита, для которого запущен pipeline. Такой тег помогает связать Docker-образ с конкретной версией кода.

Для локальной проверки используется тег:

```text
techip-tasks:local
```

## Secrets

Секреты нужны для данных, которые нельзя хранить в репозитории:

- логин registry;
- пароль или token registry;
- SSH-ключи;
- deploy tokens.

В GitHub Actions такие значения должны храниться в `Settings` → `Secrets and variables` → `Actions`.

Примеры имён secrets:

```text
REGISTRY_USERNAME
REGISTRY_PASSWORD
SSH_PRIVATE_KEY
```

## Опциональная публикация Docker-образа

Реальная публикация образа не включена, потому что registry не настроен. После настройки registry можно добавить такие шаги в `docker-build`:

```yaml
- name: Login to registry
  run: echo "${{ secrets.REGISTRY_PASSWORD }}" | docker login -u "${{ secrets.REGISTRY_USERNAME }}" --password-stdin ghcr.io

- name: Build Docker image for registry
  run: docker build -t ghcr.io/my-org/techip-tasks:${{ github.sha }} .
  working-directory: ./services/tasks

- name: Push Docker image
  run: docker push ghcr.io/my-org/techip-tasks:${{ github.sha }}
```

Перед использованием нужно заменить `my-org` на свой namespace или организацию и добавить secrets в настройках GitHub.

## Ответы на контрольные вопросы

1. **Чем CI отличается от CD?**

   CI — это Continuous Integration, то есть непрерывная интеграция. Она отвечает за автоматическую проверку проекта после изменений: установку зависимостей, запуск тестов и сборку приложения.

   CD — это Continuous Delivery или Continuous Deployment. Этот процесс связан уже не только с проверкой кода, но и с доставкой результата: публикацией Docker-образа, подготовкой релиза или автоматическим деплоем приложения.

2. **Почему pipeline должен запускать тесты?**

   Pipeline должен запускать тесты, чтобы автоматически проверять, не сломали ли новые изменения существующую функциональность. Это особенно важно в командной разработке, где код регулярно обновляется и ошибки лучше находить до деплоя.

3. **Зачем нужен автоматический build?**

   Автоматический build показывает, что проект действительно собирается в чистом окружении CI. Если приложение собирается только на компьютере разработчика, это не гарантирует, что оно соберётся на другой машине или сервере.

4. **Почему важно собирать Docker-образ в CI, а не только локально?**

   Сборка Docker-образа в CI подтверждает, что Dockerfile корректный, build context выбран правильно, зависимости доступны, а приложение можно упаковать в контейнер воспроизводимо. Это снижает риск ситуации, когда локально образ собирается, а на сервере или у другого разработчика — нет.

5. **Что такое CI secrets?**

   CI secrets — это защищённые переменные, которые хранят токены, пароли, ключи и другие чувствительные данные внутри CI-системы. Например, в GitHub Actions это могут быть `REGISTRY_USERNAME`, `REGISTRY_PASSWORD` или `SSH_PRIVATE_KEY`.

6. **Почему нельзя хранить токены и SSH-ключи в репозитории?**   

   Токены и SSH-ключи дают доступ к внешним сервисам, registry или серверам. Если закоммитить их в репозиторий, особенно публичный, ими сможет воспользоваться посторонний человек. Поэтому такие данные нужно хранить только в secrets или в защищённых настройках CI/CD.

7. **Для чего нужен тег Docker-образа?**

   Тег Docker-образа нужен для идентификации конкретной версии образа. Например, можно использовать `latest`, короткий hash коммита или полный `${{ github.sha }}`. Это позволяет понять, какая именно версия приложения была собрана и при необходимости откатиться к нужной версии.

8. **Что делает job `docker-build`?**

   Job `docker-build` собирает Docker-образ приложения внутри CI. Обычно он запускается после успешного job с тестами и сборкой кода. В этой работе `docker-build` проверяет, что сервис `tasks` можно успешно упаковать в Docker-образ.

9. **Почему в multi-service проекте важен `working-directory`?**

   В multi-service проекте `go.mod`, Dockerfile и исходный код могут находиться не в корне репозитория, а внутри отдельной папки сервиса, например `services/tasks`. Если не указать правильный `working-directory`, команды `go test`, `go build` или `docker build` могут запускаться не из той директории и завершаться ошибкой.

10. **Какие риски возникают при полностью автоматическом деплое?**

    При полностью автоматическом деплое ошибка в коде, Dockerfile, конфигурации или pipeline может сразу попасть на сервер. Это может привести к недоступности приложения, запуску неправильной версии или проблемам с конфигурацией. Поэтому для production-среды обычно добавляют проверки, ручное подтверждение, staging-окружение и возможность быстрого отката.
