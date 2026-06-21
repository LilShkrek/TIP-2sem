# Практическая работа №16. Публикация приложения в Kubernetes

## Цель

Цель работы — подготовить контейнеризированный backend-сервис `tasks` и описать его запуск в Kubernetes через ConfigMap, Deployment и Service.

## Краткая теория

Kubernetes — система оркестрации контейнеров. Она запускает приложения в контейнерах, поддерживает нужное количество экземпляров, перезапускает сбойные контейнеры и отделяет конфигурацию от образа.

Pod — минимальная единица запуска в Kubernetes. В этой работе один Pod содержит один контейнер с сервисом `tasks`.

Deployment — объект, который описывает желаемое состояние приложения: образ, число реплик, labels, probes и параметры контейнера. Если Pod завершится, Deployment создаст новый.

Service — стабильная точка доступа к Pod. В работе используется `ClusterIP`, а внешний доступ для проверки выполняется через `kubectl port-forward`.

ConfigMap — хранилище несекретной конфигурации. Здесь через него передаются `TASKS_HOST`, `TASKS_PORT`, `AUTH_BASE_URL` и `LOG_LEVEL`. Пароли, токены и другие секреты в ConfigMap не добавляются.

## Структура проекта

```text
pz16-kubernetes/
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
│   └── k8s/
│       ├── configmap.yaml
│       ├── deployment.yaml
│       └── service.yaml
├── Dockerfile
├── .dockerignore
├── go.mod
├── go.sum
└── README.md
```

## Локальный запуск Go-сервиса

По умолчанию сервис слушает `0.0.0.0:8082`. Значение `0.0.0.0` важно для запуска в контейнере и доступа через Kubernetes Service.

```bash
cd pz16-kubernetes
go run ./cmd/tasks
```

Можно переопределить конфигурацию:

```bash
TASKS_HOST=127.0.0.1 TASKS_PORT=8082 AUTH_BASE_URL=http://auth:8081 LOG_LEVEL=info go run ./cmd/tasks
```

Проверка `/health`:

```bash
curl -i http://localhost:8082/health
```

Ожидаемый ответ:

```json
{
  "status": "ok"
}
```

## Сборка Docker-образа

```bash
docker build -t techip-tasks:0.1 .
```

Фиксированный тег `techip-tasks:0.1` используется вместо `latest`, чтобы было понятно, какая версия приложения запускается в кластере, и чтобы развёртывание было воспроизводимым.

Локальная проверка контейнера:

```bash
docker run --rm -p 8082:8082 techip-tasks:0.1
```

В другом терминале:

```bash
curl -i http://localhost:8082/health
```

## Запуск локального Kubernetes-кластера

Кластер можно запустить через `kind`, `minikube`, `k3s` или учебный кластер кафедры. Перед применением манифестов проверьте подключение:

Для `kind`:

```bash
kind create cluster --name pz16
```

Для `minikube`:

```bash
minikube start
```

```bash
kubectl cluster-info
kubectl get nodes
```

## Как образ становится доступен кластеру

Kubernetes должен иметь доступ к образу `techip-tasks:0.1`. Для локального кластера образ нужно загрузить внутрь кластера или опубликовать в registry.

Для `kind`:

```bash
kind load docker-image techip-tasks:0.1
```

Для `minikube`:

```bash
minikube image load techip-tasks:0.1
```

В учебном кластере обычно требуется загрузить образ в доступный registry и указать соответствующее имя образа в `deploy/k8s/deployment.yaml`.

## Применение манифестов

```bash
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

## Проверка ресурсов

```bash
kubectl get pods
kubectl get deployment
kubectl get svc
kubectl describe deployment tasks
kubectl describe svc tasks
kubectl logs <pod-name>
```

Проверка доступа через `port-forward`:

```bash
kubectl port-forward svc/tasks 8082:8082
curl -i http://localhost:8082/health
```

## Масштабирование

Увеличить количество реплик до двух:

```bash
kubectl scale deployment tasks --replicas=2
kubectl get pods
```

Вернуть одну реплику:

```bash
kubectl scale deployment tasks --replicas=1
```

## Удаление ресурсов

```bash
kubectl delete -f deploy/k8s/service.yaml
kubectl delete -f deploy/k8s/deployment.yaml
kubectl delete -f deploy/k8s/configmap.yaml
```

## Проверки

Команды для локальной проверки проекта:

```bash
gofmt -w .
go mod tidy
go test ./...
go vet ./...
go build ./cmd/tasks
docker build -t techip-tasks:0.1 .
```

Если установлен `kubectl`, можно выполнить клиентскую проверку манифестов:

```bash
kubectl apply --dry-run=client -f deploy/k8s/configmap.yaml
kubectl apply --dry-run=client -f deploy/k8s/deployment.yaml
kubectl apply --dry-run=client -f deploy/k8s/service.yaml
```

В текущей среде `kubectl` не установлен, поэтому реальные `kubectl apply`, состояние Pod, readiness/liveness probes, Service и `port-forward` проверяются пользователем вручную в рабочем Kubernetes-кластере.

## Ответы на контрольные вопросы

1. **Что такое Kubernetes и для чего он используется?**

   Kubernetes — это система оркестрации контейнеров. Она используется для автоматического запуска, масштабирования, перезапуска и управления контейнеризированными приложениями. Kubernetes поддерживает необходимое число экземпляров сервиса и упрощает его публикацию в разных средах.

2. **Чем Pod отличается от Deployment?**

   Pod — это минимальная единица запуска в Kubernetes, содержащая один или несколько контейнеров.

   Deployment — это объект более высокого уровня, который управляет Pod: создаёт их, поддерживает заданное количество реплик, выполняет обновление и восстанавливает Pod после сбоя.

3. **Почему приложение в Kubernetes обычно публикуют через Deployment, а не через одиночный Pod?**

   Одиночный Pod не восстанавливается автоматически после удаления или сбоя. Deployment следит за количеством Pod и при необходимости создаёт новый экземпляр. Также Deployment позволяет выполнять масштабирование и управляемое обновление приложения.

4. **Зачем нужен Service и почему нельзя строить обращение к приложению напрямую через Pod?**

   Pod может быть удалён и создан заново, из-за чего его IP-адрес изменится. Service предоставляет стабильное имя и адрес для обращения к группе Pod. Он направляет трафик на Pod, соответствующие указанному selector.

5. **Что такое ConfigMap?**

   ConfigMap — это объект Kubernetes для хранения несекретной конфигурации приложения. В нём можно хранить порт, адреса других сервисов, уровень логирования и другие параметры, которые затем передаются контейнеру как переменные окружения или файлы.

6. **Чем ConfigMap отличается от Secret?**

   ConfigMap предназначен для несекретных настроек. Secret используется для конфиденциальных данных, например паролей, токенов и ключей.

   При этом Secret не гарантирует автоматическое шифрование данных во всех конфигурациях кластера. Для надёжной защиты необходимо дополнительно настраивать права доступа и шифрование хранилища Kubernetes.

7. **Для чего используется readiness probe?**

   Readiness probe проверяет, готово ли приложение принимать запросы. Пока проверка не проходит, Kubernetes не направляет трафик на Pod через Service. Это позволяет не отправлять запросы в контейнер, который ещё запускается или временно не готов к работе.

8. **Для чего используется liveness probe?**

   Liveness probe проверяет, продолжает ли приложение работать корректно. Если проверка несколько раз завершается неуспешно, Kubernetes перезапускает контейнер. Это позволяет автоматически восстанавливать зависшие или неисправные приложения.

9. **Почему важно использовать фиксированный тег образа, а не только `latest`?**

   Фиксированный тег, например `techip-tasks:0.1`, однозначно определяет версию приложения. Это делает развёртывание воспроизводимым и упрощает обновление и откат.

   Тег `latest` не показывает конкретную версию образа и может указывать на разное содержимое в разное время.

10. **Зачем нужен `kubectl port-forward`?**

    `kubectl port-forward` временно перенаправляет локальный порт компьютера на порт Pod или Service в Kubernetes-кластере. Это позволяет проверить приложение без создания внешнего LoadBalancer, NodePort или Ingress.

    Например:

    ```bash
    kubectl port-forward svc/tasks 8082:8082
    ```

    После этого сервис доступен по адресу:

    ```text
    http://localhost:8082
    ```

11. **Что делает команда `kubectl scale deployment ...`?**

    Команда изменяет требуемое количество реплик Deployment. Например:

    ```bash
    kubectl scale deployment tasks --replicas=2
    ```

    После выполнения Kubernetes создаст два Pod приложения. При уменьшении числа реплик лишние Pod будут остановлены.

12. **Почему публикация приложения в Kubernetes считается декларативной?**

    В Kubernetes разработчик описывает желаемое состояние системы в YAML-манифестах: используемый образ, количество реплик, переменные окружения, порты и проверки состояния.

    Kubernetes самостоятельно сравнивает текущее состояние с желаемым и выполняет необходимые действия. Например, если один Pod завершится, Kubernetes создаст новый, чтобы снова получить указанное количество реплик.
