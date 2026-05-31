# Практическое занятие №9

## Реализация распределённого кэша (Redis cluster)

## Цель

Освоить внедрение распределённого кэша в backend-приложение на Go и реализовать стратегию cache-aside с использованием Redis, корректного TTL, jitter и устойчивого поведения сервиса при недоступности кэша.

## Задачи

1. Изучить назначение кэширования в серверных приложениях.
2. Понять, когда кэширование действительно ускоряет систему.
3. Освоить роль Redis как внешней инфраструктурной зависимости.
4. Реализовать стратегию cache-aside для чтения сущности по идентификатору.
5. Научиться формировать ключи кэша по понятной и стабильной схеме.
6. Освоить использование TTL и случайного разброса времени жизни ключа.
7. Реализовать инвалидацию кэша при изменении и удалении данных.
8. Обеспечить деградацию сервиса при недоступности Redis без отказа основного API.
9. Научиться проверять hit, miss и fallback-сценарии на учебном стенде.

## Теоретическая часть

### Зачем backend-приложению нужен кэш

Когда серверное приложение получает запрос на чтение данных, оно чаще всего обращается к базе данных. Если запросов много, а часть данных запрашивается повторно, БД начинает выполнять лишнюю работу. В такой ситуации кэш позволяет временно хранить уже полученный результат и быстрее отдавать его при повторном обращении.

Кэширование особенно полезно, когда:

- данные читаются чаще, чем изменяются;
- чтение из БД сравнительно дороже, чем чтение из памяти;
- одни и те же сущности часто запрашиваются повторно;
- нужно снизить нагрузку на основное хранилище.

В этой работе Redis рассматривается именно как ускоритель чтения.

### Почему Redis подходит для кэширования

Redis — это очень быстрый in-memory key-value store. Его удобно использовать как внешний кэш, потому что:

- доступ к данным быстрый;
- можно задавать TTL;
- можно удалять и обновлять ключи выборочно;
- Redis легко поднять локально в Docker;
- он хорошо подходит для cache-aside сценариев.

Для учебной практики Redis удобен ещё и тем, что его поведение легко наблюдать: можно увидеть hit, miss, set, del и момент истечения TTL.

### Что такое распределённый кэш

Распределённым называют кэш, который существует не внутри одного процесса приложения, а как отдельная инфраструктурная система, доступная по сети. В отличие от обычной map в памяти приложения, такой кэш:

- не привязан к одному экземпляру сервиса;
- может использоваться несколькими экземплярами приложения;
- может быть развернут как отдельный узел или кластер.

В данной работе используется учебный стенд Redis cluster или приближённая конфигурация. Это позволяет воспринимать Redis не как вспомогательную локальную переменную, а как внешнюю зависимость backend-системы.

### Что именно кэшируется в этой работе

Минимальный обязательный сценарий — кэширование чтения задачи по идентификатору:

```text
GET /v1/tasks/{id}
```

Рекомендуемый ключ:

```text
tasks:task:<id>
```

Например:

```text
tasks:task:15
```

Для первого знакомства с Redis лучше кэшировать именно чтение одной сущности, а не списка. Такой сценарий проще для понимания и проще для инвалидации.

Опционально можно кэшировать и список задач:

```text
GET /v1/tasks
```

Тогда возможны ключи вида:

```text
tasks:list
```

или, если есть параметры:

```text
tasks:list:page=1:limit=10
```

Но список всегда сложнее в сопровождении, потому что при любой модификации данных возникает риск устаревания.

### Что такое cache-aside

Стратегия cache-aside — это один из самых распространённых подходов к кэшированию чтения.

Смысл алгоритма:

1. Сначала приложение пытается получить данные из кэша.
2. Если данные найдены — они возвращаются клиенту.
3. Если данных нет — приложение идёт в БД.
4. Если БД вернула данные — приложение кладёт их в кэш.
5. Затем ответ возвращается клиенту.

Это означает, что приложение само управляет наполнением кэша. Redis не подменяет БД и не знает бизнес-логику приложения. Именно сервис решает, когда читать из кэша, когда обращаться в БД и когда обновлять ключ.

### Почему Redis не является источником истины

Очень важно понимать архитектурную роль Redis.

Redis в данном ПЗ:

- не хранит канонические данные;
- не определяет, существует ли задача на самом деле;
- не гарантирует долговременное хранение;
- не должен ломать API при недоступности.

Источником истины остаётся БД или основной репозиторий. Redis нужен только для ускорения повторного чтения.

Это принципиальный момент. Если Redis внезапно недоступен, сервис всё равно должен обслуживать запросы через основное хранилище.

### Что такое TTL

TTL — это time to live, то есть время жизни ключа в кэше.

Если TTL не использовать, данные могут храниться бесконечно долго, а значит:

- кэш переполняется;
- данные устаревают;
- приложение всё сильнее зависит от ручной инвалидации.

Для учебной работы разумны такие значения:

- для одной сущности: 60–300 секунд;
- для списка: обычно меньше, если список часто меняется.

TTL позволяет кэшу самоочищаться и со временем обновлять содержимое.

### Что такое jitter и зачем он нужен

Если всем ключам назначить одинаковый TTL, то множество записей может истечь почти одновременно. Тогда большое число запросов резко пойдёт не в Redis, а в БД. Это создаёт всплеск нагрузки.

Чтобы уменьшить этот эффект, к базовому TTL добавляют небольшой случайный разброс — jitter.

Пример:

- базовый TTL = 120 секунд;
- jitter = от 0 до 30 секунд;
- итоговый TTL = 120 + случайное число от 0 до 30.

Так ключи истекают не одновременно, а более равномерно.

### Инвалидация кэша при изменении данных

Если данные изменились, старый кэш становится потенциально устаревшим.

Для сущности task минимально правильная политика такая:

- после `PATCH /v1/tasks/{id}` удалить ключ `tasks:task:<id>`;
- после `DELETE /v1/tasks/{id}` удалить ключ `tasks:task:<id>`;
- если кэшируется список, то после `POST`, `PATCH`, `DELETE` также удалять ключ списка.

Это простой и понятный принцип:

**изменил данные — сбросил связанный кэш**.

### Деградация при недоступности Redis

Redis — это внешняя зависимость, а значит он может быть:

- недоступен;
- остановлен;
- перегружен;
- перезапущен;
- медленным.

Требование этой практической работы: если Redis не отвечает, сервис не должен завершаться ошибкой только из-за этого.

Правильное поведение:

- ошибка Redis логируется;
- сервис идёт в БД;
- клиент получает ответ, если БД доступна.

Именно это и называется деградацией без отказа основного функционала.

## Практическая часть

### Общая идея проекта

В этой работе используется сервис tasks, в котором чтение задачи по ID будет кэшироваться в Redis.

Основной сценарий:

- первый запрос идёт в БД;
- результат записывается в Redis;
- второй запрос получает ответ из кэша;
- при изменении задачи ключ кэша удаляется;
- при остановке Redis сервис продолжает отвечать через БД.

### Рекомендуемая структура проекта

```text
pz9-redis-cache/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── cache/
│   │   ├── redis.go
│   │   ├── keys.go
│   │   └── ttl.go
│   ├── config/
│   │   └── config.go
│   ├── httpapi/
│   │   └── handler.go
│   ├── service/
│   │   └── task_service.go
│   └── task/
│       ├── model.go
│       └── repo.go
├── deploy/
│   └── redis/
│       └── docker-compose.yml
└── go.mod
```

### Шаг 1. Подготовка Redis-стенда

Есть два допустимых варианта:

**Вариант A — упрощённый:** один Redis в docker compose.

**Вариант B — предпочтительный:** учебный стенд Redis cluster.

#### Пример упрощённого docker-compose

Создайте файл `deploy/redis/docker-compose.yml`:

```yaml
version: "3.9"

services:
  redis:
    image: redis:7.4
    container_name: redis_cache
    ports:
      - "6379:6379"
```

Запуск:

```bash
cd deploy/redis
docker compose up -d
docker compose ps
```

Если Redis запущен корректно, он будет доступен на `localhost:6379`.

### Шаг 2. Подключение Redis-клиента в Go

Установите зависимость:

```bash
go get github.com/redis/go-redis/v9
```

В этой практике можно использовать современный клиент Redis для Go с поддержкой контекстов и таймаутов.

### Шаг 3. Добавление конфигурации приложения

Создайте файл `internal/config/config.go`:

```go
package config

import "time"

type Config struct {
	RedisAddr        string
	RedisPassword    string
	CacheTTL         time.Duration
	CacheTTLJitter   time.Duration
	RedisDialTimeout time.Duration
	RedisReadTimeout time.Duration
	RedisWriteTimeout time.Duration
}

func New() Config {
	return Config{
		RedisAddr:         "localhost:6379",
		RedisPassword:     "",
		CacheTTL:          120 * time.Second,
		CacheTTLJitter:    30 * time.Second,
		RedisDialTimeout:  2 * time.Second,
		RedisReadTimeout:  2 * time.Second,
		RedisWriteTimeout: 2 * time.Second,
	}
}
```

#### Пояснение

Здесь задаются:

- адрес Redis;
- базовый TTL;
- jitter;
- таймауты для работы с Redis.

Таймауты важны, потому что сервис не должен зависать при недоступности кэша.

### Шаг 4. Создание модели задачи

Создайте файл `internal/task/model.go`:

```go
package task

import "time"

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
}
```

### Шаг 5. Репозиторий с тестовыми данными

Для учебного ПЗ можно использовать in-memory репозиторий или реальную БД.

Создайте файл `internal/task/repo.go`:

```go
package task

import (
	"errors"
	"time"
)

var ErrTaskNotFound = errors.New("task not found")

type Repo struct {
	data map[int64]Task
}

func NewRepo() *Repo {
	return &Repo{
		data: map[int64]Task{
			1: {
				ID:          1,
				Title:       "Изучить Redis",
				Description: "Разобрать cache-aside",
				DueDate:     time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
			},
			2: {
				ID:          2,
				Title:       "Сделать ПЗ",
				Description: "Реализовать кэширование по id",
				DueDate:     time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

func (r *Repo) GetByID(id int64) (Task, error) {
	t, ok := r.data[id]
	if !ok {
		return Task{}, ErrTaskNotFound
	}

	return t, nil
}

func (r *Repo) Update(task Task) error {
	if _, ok := r.data[task.ID]; !ok {
		return ErrTaskNotFound
	}

	r.data[task.ID] = task
	return nil
}

func (r *Repo) Delete(id int64) error {
	if _, ok := r.data[id]; !ok {
		return ErrTaskNotFound
	}

	delete(r.data, id)
	return nil
}
```

### Шаг 6. Создание Redis-клиента

Создайте файл `internal/cache/redis.go`:

```go
package cache

import (
	"context"
	"time"

	"example.com/pz9-redis-cache/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           0,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
	})
}

func Ping(ctx context.Context, client *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}
```

### Шаг 7. Формирование ключей кэша

Создайте файл `internal/cache/keys.go`:

```go
package cache

import "fmt"

func TaskByIDKey(id int64) string {
	return fmt.Sprintf("tasks:task:%d", id)
}

func TasksListKey() string {
	return "tasks:list"
}
```

#### Пояснение

Ключи должны быть:

- понятными;
- предсказуемыми;
- единообразными.

Схема `tasks:task:<id>` как раз подходит под эту задачу и соответствует исходной логике вашего черновика.

### Шаг 8. Расчёт TTL с jitter

Создайте файл `internal/cache/ttl.go`:

```go
package cache

import (
	"math/rand"
	"time"
)

func TTLWithJitter(base time.Duration, jitter time.Duration) time.Duration {
	if jitter <= 0 {
		return base
	}

	extra := time.Duration(rand.Int63n(int64(jitter) + 1))
	return base + extra
}
```

#### Пояснение

Эта функция позволяет получить итоговый TTL, который немного отличается для разных ключей. Именно это и снижает риск одновременного истечения большого числа записей.

### Шаг 9. Реализация сервисного слоя с cache-aside

Создайте файл `internal/service/task_service.go`:

```go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"example.com/pz9-redis-cache/internal/cache"
	"example.com/pz9-redis-cache/internal/config"
	"example.com/pz9-redis-cache/internal/task"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	repo  *task.Repo
	redis *redis.Client
	cfg   config.Config
}

func NewTaskService(repo *task.Repo, redisClient *redis.Client, cfg config.Config) *TaskService {
	return &TaskService{
		repo:  repo,
		redis: redisClient,
		cfg:   cfg,
	}
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int64) (task.Task, error) {
	key := cache.TaskByIDKey(id)

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var t task.Task
			if err := json.Unmarshal([]byte(cached), &t); err == nil {
				log.Println("cache hit:", key)
				return t, nil
			}

			log.Println("cache decode error:", err)
		} else if !errors.Is(err, redis.Nil) {
			log.Println("redis read error:", err)
		} else {
			log.Println("cache miss:", key)
		}
	}

	t, err := s.repo.GetByID(id)
	if err != nil {
		return task.Task{}, err
	}

	if s.redis != nil {
		bytes, err := json.Marshal(t)
		if err != nil {
			log.Println("cache encode error:", err)
			return t, nil
		}

		ttl := cache.TTLWithJitter(s.cfg.CacheTTL, s.cfg.CacheTTLJitter)
		if err := s.redis.Set(ctx, key, bytes, ttl).Err(); err != nil {
			log.Println("redis write error:", err)
		}
	}

	return t, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, t task.Task) error {
	if err := s.repo.Update(t); err != nil {
		return err
	}

	if s.redis != nil {
		key := cache.TaskByIDKey(t.ID)
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			log.Println("redis delete error:", err)
		}
	}

	return nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if s.redis != nil {
		key := cache.TaskByIDKey(id)
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			log.Println("redis delete error:", err)
		}
	}

	return nil
}
```

#### Что важно в этой реализации

Здесь реализованы все ключевые требования:

- сначала читаем из Redis;
- при miss идём в БД;
- после чтения из БД кладём объект в кэш;
- используем TTL и jitter;
- при ошибках Redis не падаем;
- при update/delete инвалидируем кэш.

Именно такое поведение и соответствует вашему базовому сценарию из текущего файла ПЗ №9.

### Шаг 10. HTTP-обработчик

Создайте файл `internal/httpapi/handler.go`:

```go
package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/pz9-redis-cache/internal/service"
	"example.com/pz9-redis-cache/internal/task"
)

type Handler struct {
	service *service.TaskService
}

func NewHandler(service *service.TaskService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawID := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	t, err := h.service.GetTaskByID(r.Context(), id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(t)
}

func (h *Handler) PatchTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var t task.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateTask(r.Context(), t); err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawID := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

### Шаг 11. Точка входа приложения

Создайте файл `cmd/server/main.go`:

```go
package main

import (
	"context"
	"log"
	"net/http"

	"example.com/pz9-redis-cache/internal/cache"
	"example.com/pz9-redis-cache/internal/config"
	"example.com/pz9-redis-cache/internal/httpapi"
	"example.com/pz9-redis-cache/internal/service"
	"example.com/pz9-redis-cache/internal/task"
)

func main() {
	cfg := config.New()
	repo := task.NewRepo()
	redisClient := cache.NewRedisClient(cfg)

	if err := cache.Ping(context.Background(), redisClient); err != nil {
		log.Println("warning: redis is unavailable at startup:", err)
	}

	taskService := service.NewTaskService(repo, redisClient, cfg)
	handler := httpapi.NewHandler(taskService)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetTaskByID(w, r)
		case http.MethodPatch:
			handler.PatchTask(w, r)
		case http.MethodDelete:
			handler.DeleteTask(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("server started on :8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatal(err)
	}
}
```

### Шаг 12. Проверка чтения и заполнения кэша

Запустите Redis:

```bash
cd deploy/redis
docker compose up -d
```

Запустите сервис:

```bash
go run ./cmd/server
```

Выполните первый запрос:

```bash
curl http://localhost:8082/v1/tasks/1
```

Затем сразу второй:

```bash
curl http://localhost:8082/v1/tasks/1
```

#### Что должно произойти

Первый запрос:

- не находит ключ в Redis;
- идёт в репозиторий;
- получает задачу;
- кладёт её в кэш.

Второй запрос:

- должен вернуть данные из Redis.

Если в логах добавлены сообщения `cache miss` и `cache hit`, это будет видно особенно наглядно.

### Шаг 13. Проверка инвалидации при изменении

Выполните обновление задачи:

```bash
curl -X PATCH http://localhost:8082/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"id":1,"title":"Обновлённая задача","description":"Новый текст","due_date":"2026-01-22T00:00:00Z"}'
```

После этого снова выполните:

```bash
curl http://localhost:8082/v1/tasks/1
```

#### Что должно произойти

После PATCH ключ `tasks:task:1` должен быть удалён.

Следующий GET снова пойдёт в репозиторий и заново заполнит кэш уже свежими данными.

### Шаг 14. Проверка удаления

Выполните:

```bash
curl -X DELETE http://localhost:8082/v1/tasks/1
```

После этого:

```bash
curl http://localhost:8082/v1/tasks/1
```

Ожидаемый результат — `404 Not Found`.

Если до удаления в Redis был кэшированный объект, он тоже должен быть удалён.

### Шаг 15. Проверка деградации при остановке Redis

Остановите Redis:

```bash
cd deploy/redis
docker compose stop
```

После этого снова выполните:

```bash
curl http://localhost:8082/v1/tasks/2
```

#### Ожидаемое поведение

Сервис:

- не должен упасть;
- не должен зависнуть надолго;
- должен получить данные из репозитория;
- должен записать предупреждение об ошибке Redis.

Это обязательная часть работы, и она прямо заложена в вашем текущем варианте ПЗ №9 как одно из ключевых требований.

### Шаг 16. Что важно понять по итогам практики

После выполнения этой работы вы должны чётко понимать:

- Redis ускоряет чтение, но не заменяет БД;
- кэш нельзя считать единственным местом хранения данных;
- cache-aside — это не магия Redis, а прикладной алгоритм в коде сервиса;
- TTL нужен для самообновления кэша;
- jitter нужен для более равномерного истечения ключей;
- при модификации данных кэш должен инвалидироваться;
- при отказе Redis приложение должно продолжать работать через основное хранилище.

## Дополнительные задания

### Вариант 1. Кэширование списка задач

Реализуйте кэширование маршрута:

```text
GET /v1/tasks
```

с ключом:

```text
tasks:list
```

или с учётом параметров пагинации.

### Вариант 2. Отдельное логирование hit/miss

Добавьте в сервисный слой явное логирование:

- cache hit;
- cache miss;
- cache set;
- cache invalidated.

### Вариант 3. Разделить key-builder и serializer

Вынесите отдельно:

- генерацию ключей;
- сериализацию в JSON;
- десериализацию из JSON.

### Вариант 4. Реализовать отрицательное кэширование

Можно рассмотреть вариант временного кэширования факта отсутствия сущности, но только как дополнительное задание и с аккуратным TTL.

## Типичные ошибки

### Ошибка 1. Redis недоступен, и сервис падает

Это противоречит цели работы. Redis должен быть ускорителем, а не причиной отказа основного API.

### Ошибка 2. Ключи формируются хаотично

Нужно использовать единый формат, например `tasks:task:<id>`.

### Ошибка 3. Нет TTL

Тогда кэш не самообновляется и может надолго устареть.

### Ошибка 4. Нет jitter

При большом числе ключей это может приводить к всплескам нагрузки при одновременном истечении записей.

### Ошибка 5. После PATCH или DELETE кэш не сбрасывается

Тогда клиент может получать устаревшие данные.

### Ошибка 6. При ошибке чтения JSON из кэша сервис отдаёт 500

Правильнее считать такой случай miss и пойти в БД.

### Ошибка 7. Нет таймаутов на Redis

Тогда приложение может подвисать при недоступности кэша.

## Контрольные вопросы

1. Что такое cache-aside?
2. Почему Redis не должен быть источником истины?
3. Зачем нужен TTL?
4. Что такое jitter?
5. Почему одинаковый TTL для всех ключей может быть проблемой?
6. Как должен вести себя сервис при недоступности Redis?
7. Почему кэш нужно инвалидировать после изменения данных?
8. Чем кэширование одной сущности проще, чем кэширование списка?
9. В чём смысл ключа вида `tasks:task:<id>`?
10. Почему Redis рассматривается как внешняя инфраструктурная зависимость?