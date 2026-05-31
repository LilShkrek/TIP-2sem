# Практическая работа №4

## Цель

Освоить базовую организацию мониторинга backend-приложения на Go с использованием Prometheus для сбора метрик и Grafana для визуализации.

## Задание

Реализовать учебное Go-приложение с HTTP-маршрутами:

- `GET /health`
- `GET /students/{id}`
- `GET /metrics`

Приложение должно публиковать Prometheus-метрики:

- `app_http_requests_total` — общее число HTTP-запросов;
- `app_http_errors_total` — число ответов с ошибочными HTTP-статусами;
- `app_http_request_duration_seconds` — длительность обработки HTTP-запросов.

## Структура проекта

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── httpapi/
│   │   ├── handler.go
│   │   ├── middleware.go
│   │   └── response_writer.go
│   ├── metrics/
│   │   └── metrics.go
│   └── student/
│       ├── model.go
│       └── repo.go
├── monitoring/
│   └── prometheus.yml
├── go.mod
├── go.sum
└── README.md
```

## Запуск приложения

Установите зависимости:

```bash
go mod tidy
```

Запустите сервер:

```bash
go run ./cmd/server
```

Приложение будет доступно на порту `8080`.

## Проверка маршрутов

```bash
curl http://localhost:8080/health
curl http://localhost:8080/students/1
curl http://localhost:8080/students/999
```

Ожидаемые результаты:

- `/health` возвращает JSON со статусом приложения;
- `/students/1` возвращает данные студента;
- `/students/999` возвращает ошибку `404`.

## Проверка `/metrics`

Сгенерируйте несколько запросов:

```bash
for i in {1..5}; do curl -s http://localhost:8080/health > /dev/null; done
for i in {1..5}; do curl -s http://localhost:8080/students/1 > /dev/null; done
for i in {1..3}; do curl -s http://localhost:8080/students/999 > /dev/null; done
```

Откройте метрики:

```bash
curl http://localhost:8080/metrics
```

В выводе должны появиться строки с метриками:

```text
app_http_requests_total
app_http_errors_total
app_http_request_duration_seconds
```

## Запуск Prometheus

Перед запуском Prometheus необходимо запустить само Go-приложение:

```bash
go run ./cmd/server
```

Приложение должно быть доступно на порту `8080`.

Проверить публикацию метрик можно командой:

```bash
curl http://localhost:8080/metrics
```

В выводе должны присутствовать метрики приложения:

```text
app_http_requests_total
app_http_errors_total
app_http_request_duration_seconds
```

Файл конфигурации Prometheus находится в каталоге `monitoring`:

```text
monitoring/prometheus.yml
```

Пример конфигурации:

```yaml
global:
  scrape_interval: 5s

scrape_configs:
  - job_name: "go_app"
    static_configs:
      - targets: ["localhost:8080"]

  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
```

Если Prometheus установлен локально, его можно запустить командой:

```bash
prometheus --config.file=monitoring/prometheus.yml
```

После запуска необходимо открыть Prometheus в браузере:

```text
http://localhost:9090
```

Затем перейти в раздел:

```text
Status -> Targets
```

В списке targets должен появиться target приложения:

```text
job="go_app", instance="localhost:8080"
```

Его состояние должно быть `UP`.

Если в Targets отображаются только `prometheus` или `node`, значит Prometheus был запущен не с конфигурацией из текущего проекта.

## Запуск Prometheus через Docker

Также Prometheus можно запустить через Docker. На Linux удобно использовать сетевой режим `host`, чтобы контейнер Prometheus мог обращаться к приложению на `localhost:8080`:

```bash
docker run --rm \
  --network host \
  -v "$PWD/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml" \
  prom/prometheus \
  --config.file=/etc/prometheus/prometheus.yml
```

После запуска Prometheus будет доступен по адресу:

```text
http://localhost:9090
```

В разделе Targets также должен быть target:

```text
go_app -> localhost:8080 -> UP
```

## Запуск Grafana

Grafana можно запустить через Docker:

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  --add-host=host.docker.internal:host-gateway \
  grafana/grafana
```

После запуска Grafana будет доступна в браузере:

```text
http://localhost:3000
```

Стандартные данные для первого входа:

```text
login: admin
password: admin
```

После входа Grafana может предложить сменить пароль. Для учебной работы можно задать любой новый пароль или пропустить этот шаг, если интерфейс позволяет.

## Подключение Prometheus к Grafana

Для подключения Prometheus как источника данных нужно выполнить следующие действия:

1. Открыть Grafana:

   ```text
   http://localhost:3000
   ```

2. Перейти в раздел:

   ```text
   Connections -> Data sources
   ```

3. Нажать:

   ```text
   Add data source
   ```

4. Выбрать тип источника данных:

   ```text
   Prometheus
   ```

5. В поле URL указать адрес Prometheus.

   Если Grafana запущена локально без Docker:

   ```text
   http://localhost:9090
   ```

   Если Grafana запущена в Docker, а Prometheus запущен на хостовой машине:

   ```text
   http://host.docker.internal:9090
   ```

6. Нажать:

   ```text
   Save & test
   ```

После успешной проверки Grafana должна показать, что источник данных Prometheus подключён.

## Создание панелей в Grafana

После подключения Prometheus можно создать dashboard:

1. Перейти в раздел:

   ```text
   Dashboards
   ```

2. Нажать:

   ```text
   Create dashboard
   ```

3. Добавить новую панель:

   ```text
   Add visualization
   ```

4. Выбрать источник данных Prometheus.

Для проверки можно создать несколько панелей с PromQL-запросами.

Общее число HTTP-запросов:

```promql
sum(app_http_requests_total)
```

Число ошибочных HTTP-запросов:

```promql
sum(app_http_errors_total)
```

Количество запросов по маршрутам:

```promql
sum by (path) (app_http_requests_total)
```

Количество ошибок по HTTP-статусам:

```promql
sum by (status_code) (app_http_errors_total)
```

Среднее время обработки HTTP-запросов:

```promql
sum(rate(app_http_request_duration_seconds_sum[1m]))
/
sum(rate(app_http_request_duration_seconds_count[1m]))
```

95-й перцентиль времени обработки запросов:

```promql
histogram_quantile(
  0.95,
  sum by (le) (rate(app_http_request_duration_seconds_bucket[1m]))
)
```

Для появления данных в Grafana нужно предварительно выполнить несколько запросов к приложению:

```bash
for i in {1..5}; do curl -s http://localhost:8080/health > /dev/null; done
for i in {1..5}; do curl -s http://localhost:8080/students/1 > /dev/null; done
for i in {1..3}; do curl -s http://localhost:8080/students/999 > /dev/null; done
```

После этого Prometheus соберёт метрики, и они станут доступны для отображения в Grafana.


## Проверка кода

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
```

## Ответы на контрольные вопросы

1. **Что такое метрики приложения?**

   Метрики приложения — это числовые показатели, которые описывают состояние сервиса и его поведение во времени. Например, количество HTTP-запросов, число ошибок, длительность обработки запросов, нагрузка или количество активных соединений.

2. **Чем метрики отличаются от логов?**

   Логи показывают, что произошло в конкретный момент: какой запрос пришёл, какая ошибка возникла, какой обработчик был вызван.  
   Метрики показывают общую картину работы приложения: сколько всего было запросов, как часто появляются ошибки, насколько быстро отвечает сервис и как эти значения меняются со временем.

3. **Какую роль выполняет Prometheus?**

   Prometheus собирает метрики с приложения, сохраняет их как временные ряды и позволяет выполнять по ним запросы через PromQL. В этой работе Prometheus периодически обращается к маршруту `/metrics` Go-приложения и получает оттуда значения метрик.

4. **Что такое scraping в Prometheus?**

   Scraping — это процесс регулярного опроса приложения со стороны Prometheus. Приложение публикует метрики на endpoint `/metrics`, а Prometheus через заданный интервал делает HTTP-запрос к этому endpoint и сохраняет полученные значения.

5. **Зачем приложению маршрут `/metrics`?**

   Маршрут `/metrics` нужен для того, чтобы приложение могло отдавать Prometheus свои метрики в понятном для него формате. Без этого endpoint Prometheus не сможет получить данные о количестве запросов, ошибках и времени обработки.

6. **Что делает `promhttp.Handler()`?**

   `promhttp.Handler()` создаёт HTTP-обработчик, который отдаёт зарегистрированные Prometheus-метрики. В этой работе он подключается к маршруту `/metrics`, чтобы Prometheus мог считывать метрики приложения.

7. **Для чего нужна Grafana?**

   Grafana нужна для визуализации метрик. Prometheus собирает и хранит данные, а Grafana подключается к нему как к источнику данных и строит панели, графики и дашборды. Так проще анализировать количество запросов, ошибки и задержки.

8. **Какие три основные метрики реализованы в этой работе?**

   В работе реализованы три основные метрики:

    - `app_http_requests_total` — общее количество HTTP-запросов;
    - `app_http_errors_total` — количество HTTP-ответов с ошибочными статусами;
    - `app_http_request_duration_seconds` — длительность обработки HTTP-запросов.

9. **Что показывает Histogram?**

   Histogram показывает распределение значений по диапазонам. В этой работе Histogram используется для измерения длительности HTTP-запросов. Благодаря этому можно анализировать не только среднее время ответа, но и, например, 95-й перцентиль задержки.

10. **Почему мониторинг важен для сопровождения backend-приложений?**

    Мониторинг позволяет быстро видеть состояние приложения и замечать проблемы до того, как они станут критичными. По метрикам можно понять, растёт ли число ошибок, увеличивается ли время ответа, доступно ли приложение и как оно ведёт себя под нагрузкой. Это важно для эксплуатации и поддержки backend-сервисов.