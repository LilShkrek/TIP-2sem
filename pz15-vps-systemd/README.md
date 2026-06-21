# Практическая работа №15. Деплой Go-сервиса на VPS через systemd

## Цель

Реализовать небольшой Go-сервис `tasks`, подготовить его к сборке в Linux-бинарник и описать ручной деплой на VPS через `systemd`.

VPS — это виртуальный Linux-сервер, на котором можно постоянно запускать backend-приложения. `systemd` управляет такими приложениями как службами: запускает их при старте системы, перезапускает при сбоях, показывает статус и отдаёт логи через `journalctl`.

Реальный деплой на VPS в рамках локальной проверки не выполнялся. Все действия с VPS выполняются пользователем вручную.

## Структура проекта

```text
pz15-vps-systemd/
├── cmd/
│   └── tasks/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   └── httpserver/
│       ├── server.go
│       └── server_test.go
├── deploy/
│   ├── env/
│   │   └── tasks.env.example
│   └── systemd/
│       └── tasks.service
├── scripts/
│   ├── build-linux.sh
│   ├── install-vps.sh
│   ├── rollback-vps.sh
│   └── update-vps.sh
├── go.mod
└── README.md
```

## Конфигурация

Сервис читает настройки из переменных окружения:

```env
TASKS_HOST=127.0.0.1
TASKS_PORT=8082
LOG_LEVEL=info
```

На VPS файл конфигурации должен находиться здесь:

```text
/etc/tasks/tasks.env
```

Рекомендуемые права:

```bash
sudo chown root:root /etc/tasks/tasks.env
sudo chmod 600 /etc/tasks/tasks.env
```

## Локальный запуск

```bash
cd pz15-vps-systemd
go run ./cmd/tasks
```

По умолчанию сервис слушает `127.0.0.1:8082`.

Проверка `/health`:

```bash
curl -i http://127.0.0.1:8082/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

## Сборка Linux-бинарника

```bash
./scripts/build-linux.sh
```

Скрипт по умолчанию использует `GOOS=linux` и `GOARCH=amd64`, результат сохраняется в `bin/tasks`.

Для ARM-сервера можно изменить архитектуру:

```bash
GOARCH=arm64 ./scripts/build-linux.sh
```

Собранный каталог `bin/` игнорируется Git.

## Ручной деплой на VPS

Подключение по SSH:

```bash
ssh user@<VPS_IP>
```

Копирование бинарника с локальной машины:

```bash
scp bin/tasks user@<VPS_IP>:/tmp/tasks
```

Также скопируйте пример env-файла и unit-файл:

```bash
scp deploy/env/tasks.env.example user@<VPS_IP>:/tmp/tasks.env
scp deploy/systemd/tasks.service user@<VPS_IP>:/tmp/tasks.service
```

На VPS создайте пользователя и директории:

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin tasksuser
sudo mkdir -p /opt/tasks /etc/tasks
sudo chown tasksuser:tasksuser /opt/tasks
```

Установите бинарник:

```bash
sudo mv /tmp/tasks /opt/tasks/tasks
sudo chown tasksuser:tasksuser /opt/tasks/tasks
sudo chmod 755 /opt/tasks/tasks
```

Создайте конфигурацию:

```bash
sudo mv /tmp/tasks.env /etc/tasks/tasks.env
sudo chown root:root /etc/tasks/tasks.env
sudo chmod 600 /etc/tasks/tasks.env
```

Установите `tasks.service`:

```bash
sudo mv /tmp/tasks.service /etc/systemd/system/tasks.service
sudo chown root:root /etc/systemd/system/tasks.service
sudo chmod 644 /etc/systemd/system/tasks.service
```

Перечитайте конфигурацию systemd, запустите службу и включите автозапуск:

```bash
sudo systemctl daemon-reload
sudo systemctl start tasks
sudo systemctl enable tasks
```

Проверка статуса:

```bash
sudo systemctl status tasks
```

Просмотр логов:

```bash
sudo journalctl -u tasks --no-pager -n 100
sudo journalctl -u tasks -f
```

Проверка `/health` на VPS:

```bash
curl -i http://127.0.0.1:8082/health
```

Если порт открыт наружу и firewall это разрешает:

```bash
curl -i http://<VPS_IP>:8082/health
```

## Установка через скрипт

Скрипт выполняется непосредственно на VPS с `sudo`. Он создаёт пользователя, каталоги, устанавливает бинарник, env-файл и unit-файл, делает резервные копии существующих файлов, выполняет `systemctl daemon-reload`, запускает службу и показывает статус.

Пример:

```bash
sudo TASKS_BINARY=/tmp/tasks TASKS_ENV=/tmp/tasks.env TASKS_UNIT=/tmp/tasks.service ./scripts/install-vps.sh
```

## Обновление

На локальной машине соберите новый бинарник и скопируйте его на VPS:

```bash
./scripts/build-linux.sh
scp bin/tasks user@<VPS_IP>:/tmp/tasks
```

На VPS выполните:

```bash
sudo TASKS_BINARY=/tmp/tasks ./scripts/update-vps.sh
```

Скрипт остановит службу, сохранит текущий бинарник в `/opt/tasks/tasks.old`, установит новый бинарник, запустит сервис и проверит состояние.

## Откат

Если обновление неуспешно, на VPS выполните:

```bash
sudo ./scripts/rollback-vps.sh
```

Скрипт проверит наличие `/opt/tasks/tasks.old`, остановит службу, восстановит старый бинарник, задаст владельца и права, снова запустит сервис и покажет статус.

## Основные команды управления службой

```bash
sudo systemctl start tasks
sudo systemctl stop tasks
sudo systemctl restart tasks
sudo systemctl status tasks
sudo systemctl enable tasks
sudo systemctl disable tasks
sudo journalctl -u tasks --no-pager -n 100
```

## Firewall и NGINX

В учебном варианте можно открыть порт `8082` наружу, если это разрешено правилами firewall. В промышленной конфигурации сервис лучше оставлять на `127.0.0.1`, а внешний трафик принимать через NGINX на `80/443` и проксировать к `127.0.0.1:8082`.

## Локальные проверки

Фактически выполнялись локально:

```bash
gofmt -w .
go mod tidy
go test ./...
go vet ./...
GOOS=linux GOARCH=amd64 go build -o bin/tasks ./cmd/tasks
bash -n scripts/*.sh
curl -i http://127.0.0.1:8082/health
```

Проверка реального VPS, SSH-подключение, копирование через `scp`, установка systemd-службы и проверка через публичный IP выполняются пользователем вручную.

## Ответы на контрольные вопросы

1. **Что такое VPS и зачем он нужен backend-разработчику?**

   VPS — это виртуальный выделенный сервер, работающий в дата-центре и доступный через интернет. Backend-разработчик может размещать на нём приложения, базы данных и другие сервисы, чтобы они работали постоянно и были доступны пользователям независимо от компьютера разработчика.

2. **Почему запуск приложения на VPS отличается от локального запуска на компьютере разработчика?**

   Локально приложение обычно запускается вручную из терминала или IDE. На VPS сервис должен работать постоянно, автоматически запускаться после перезагрузки, восстанавливаться после ошибок, использовать отдельную конфигурацию и предоставлять системные логи. Поэтому для запуска применяются специальные средства, например systemd.

3. **Для чего используется systemd?**

   systemd используется для управления системными службами Linux. Он позволяет запускать и останавливать приложение, включать автозапуск, автоматически перезапускать сервис после сбоя, проверять его состояние и просматривать логи.

4. **Почему не рекомендуется запускать серверное приложение от root?**

   Пользователь `root` имеет максимальные права в системе. Если приложение содержит уязвимость или будет взломано, злоумышленник сможет получить полный контроль над сервером. Поэтому сервис лучше запускать от отдельного пользователя с минимально необходимыми правами, например `tasksuser`.

5. **Зачем выносить конфигурацию в отдельный env-файл?**

   Отдельный env-файл позволяет изменять порт, адрес прослушивания, уровень логирования и другие параметры без изменения и повторной компиляции кода. Также это помогает не хранить пароли, токены и другие секреты непосредственно в исходном коде и репозитории.

6. **Что делает параметр `Restart=always`?**

   Параметр `Restart=always` указывает systemd автоматически перезапускать сервис после его завершения независимо от причины остановки. Это повышает доступность приложения при программных сбоях.

7. **Для чего нужен `EnvironmentFile` в unit-файле?**

   `EnvironmentFile` указывает systemd путь к файлу с переменными окружения, которые необходимо передать приложению при запуске. В данной работе используется файл:

   ```text
   /etc/tasks/tasks.env
   ```

   Из него сервис получает значения `TASKS_HOST`, `TASKS_PORT` и другие настройки.

8. **Как проверить состояние службы через `systemctl`?**

   Для проверки состояния службы используется команда:

   ```bash
   sudo systemctl status tasks
   ```

   Она показывает, запущен ли сервис, его идентификатор процесса, время работы и последние сообщения журнала.

9. **Как посмотреть логи сервиса через `journalctl`?**

   Для просмотра последних записей конкретного сервиса используется команда:

   ```bash
   sudo journalctl -u tasks --no-pager -n 100
   ```

   Для просмотра логов в реальном времени:

   ```bash
   sudo journalctl -u tasks -f
   ```

10. **Что нужно сделать перед обновлением unit-файла systemd?**

    После изменения unit-файла необходимо выполнить:

    ```bash
    sudo systemctl daemon-reload
    ```

    Эта команда заставляет systemd перечитать конфигурацию служб. Затем сервис нужно перезапустить:

    ```bash
    sudo systemctl restart tasks
    ```

11. **Почему полезно иметь процедуру отката версии?**

    Новая версия приложения может содержать ошибку и не запуститься после обновления. Наличие резервного бинарника позволяет быстро вернуть предыдущую рабочую версию и восстановить доступность сервиса.

12. **Зачем в реальных системах часто используют NGINX перед приложением?**

    NGINX принимает внешние HTTP- и HTTPS-запросы и перенаправляет их внутреннему приложению. Он может выполнять TLS-терминацию, балансировку нагрузки, ограничение запросов, кэширование и проксирование. При этом Go-сервис может безопасно слушать только локальный адрес `127.0.0.1`, не открывая свой порт напрямую в интернет.
