# InfluxDB + Telegraf + Grafana: Архитектура и взаимодействие

## 📋 Содержание
1. [Общая архитектура](#общая-архитектура)
2. [Компоненты системы](#компоненты-системы)
3. [Как работает InfluxDB](#как-работает-influxdb)
4. [Роль Telegraf](#роль-telegraf)
5. [Прямая запись из Go](#прямая-запись-из-go)
6. [Визуализация в Grafana](#визуализация-в-grafana)
7. [Конфигурационные файлы](#конфигурационные-файлы)

---

## Общая архитектура

```
┌─────────────────────────────────────────────────────────────────────┐
│                    SSO Service (Go)                                 │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  InfluxDB Client (прямая запись)                            │   │
│  │  - auth_operation (с тегом user_role!)                      │   │
│  │  - http_request                                             │   │
│  │  - users (total, active_last_24h, new_today)                │   │
│  └──────────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Prometheus /metrics endpoint                               │   │
│  │  - sso_http_requests_total                                  │   │
│  │  - sso_http_request_duration_seconds                        │   │
│  │  - sso_total_users                                          │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
           ↓ WriteAPI (асинхронно)              ↓ HTTP GET /metrics (каждые 10 сек)
           │                                     │
           │                              ┌──────┴─────────────────────────────────┐
           │                              │        Telegraf                         │
           │                              │  ┌──────────────────────────────────┐   │
           │                              │  │  Input: Prometheus Scraper       │   │
           │                              │  │  - Скрейпит /metrics каждые 10s │   │
           │                              │  │  - Преобразует в Line Protocol  │   │
           │                              │  └──────────────────────────────────┘   │
           │                              │  ┌──────────────────────────────────┐   │
           │                              │  │  Input: PostgreSQL               │   │
           │                              │  │  Input: Docker                   │   │
           │                              │  │  Input: CPU/Memory/Disk          │   │
           │                              │  └──────────────────────────────────┘   │
           │                              └────────┬───────────────────────────────┘
           │                                       ↓ HTTP POST (каждые 10 сек)
           └───────────────────────────────────────┘
                                     ↓
┌─────────────────────────────────────────────────────────────────────┐
│                        InfluxDB 2.x                                 │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Time Series Database (TSM storage engine)                   │   │
│  │  Organization: sso-org                                       │   │
│  │  Bucket: sso-metrics (retention 7 days)                      │   │
│  │                                                              │   │
│  │  Measurements:                                               │   │
│  │  - auth_operation (tags: user_role, operation, status)       │   │
│  │  - http_request (tags: method, path, status)                 │   │
│  │  - users (tags: environment, region)                         │   │
│  │  - prometheus (все Prometheus метрики)                       │   │
│  │  - cpu, mem, disk, net, docker, postgresql                   │   │
│  └──────────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Flux Query Engine                                           │   │
│  │  - Предоставляет /api/v2/query endpoint                      │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                          ↓ Flux запросы (каждые 5 сек)
┌─────────────────────────────────────────────────────────────────────┐
│                         Grafana                                     │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Dashboard: SSO API Metrics Dashboard (InfluxDB)             │   │
│  │  - HTTP Requests per Second                                  │   │
│  │  - HTTP Duration p95                                         │   │
│  │  - Average Response Time                                     │   │
│  │  - Total Registered Users                                    │   │
│  │  - Client/Manager/Admin Auth Success (6 gauge метрик!)       │   │
│  │  - Total HTTP Requests                                       │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                          ↓ Пользователь видит в браузере
                    http://localhost:3000
```

---

## Компоненты системы

### 1. **SSO Service (Go приложение)**

**Роль:** Записывает метрики **напрямую** в InfluxDB через Go-клиент + экспортирует Prometheus метрики.

**Ключевые файлы:**
- [`pkg/metrics/influxdb.go`](../pkg/metrics/influxdb.go) - InfluxDB writer
- [`pkg/metrics/metrics.go`](../pkg/metrics/metrics.go) - интеграция InfluxDB + Prometheus
- [`internal/application/server.go`](../internal/application/server.go) - инициализация InfluxDB клиента
- [`internal/application/auth/handler.go`](../internal/application/auth/handler.go) - запись auth_operation метрик
- [`config/config.yaml`](../config/config.yaml) - конфигурация InfluxDB

**Что записывается в InfluxDB:**

1. **auth_operation** (записывается только из Go, НЕ через Telegraf):
   ```
   Measurement: auth_operation
   Tags: operation=login, status=success, user_role=client, service=sso-api, environment=dev
   Fields: count=1, duration_ms=0
   Timestamp: 2025-11-26T15:34:52.567Z
   ```

2. **http_request** (записывается из Go):
   ```
   Measurement: http_request
   Tags: method=POST, path=/auth/logIn, status=200, service=sso-api, environment=dev
   Fields: duration_seconds=0.006, count=1
   Timestamp: 2025-11-26T15:34:52.567Z
   ```

3. **users** (записывается из Go каждые 30 сек):
   ```
   Measurement: users
   Tags: service=sso-api, environment=dev, region=local
   Fields: total=22, active_last_24h=0, new_today=0
   Timestamp: 2025-11-26T15:34:11.395Z
   ```

**Конфигурация InfluxDB:**

**Файл:** [`config/config.yaml`](../config/config.yaml)

```yaml
influxdb:
  enabled: true
  url: "http://influxdb:8086"
  token: "my-super-secret-auth-token"
  org: "sso-org"
  bucket: "sso-metrics"
```

---

### 2. **Telegraf**

**Роль:** Собирает метрики из разных источников и отправляет в InfluxDB.

**Docker контейнер:** `sso-telegraf`
**Образ:** `telegraf:1.31-alpine`

**Ключевые файлы:**
- [`deployments/telegraf/telegraf.conf`](../deployments/telegraf/telegraf.conf) - конфигурация
- [`deployments/docker-compose.yaml`](../deployments/docker-compose.yaml) - запуск контейнера

**Что собирает:**

1. **Prometheus метрики из SSO** (`[[inputs.prometheus]]`):
   - Скрейпит `http://sso:8081/metrics` каждые 10 секунд
   - Преобразует Prometheus формат в InfluxDB Line Protocol
   - Сохраняет как measurement `prometheus`

2. **PostgreSQL метрики** (`[[inputs.postgresql]]`):
   - Подключается к `host.docker.internal:5432`
   - Собирает: количество соединений, размер БД, статистику таблиц
   - Measurement: `postgresql`

3. **System метрики** (`[[inputs.cpu]]`, `[[inputs.mem]]`, `[[inputs.disk]]`):
   - CPU usage, memory, disk
   - Measurements: `cpu`, `mem`, `disk`, `net`

4. **Docker метрики** (`[[inputs.docker]]`):
   - ⚠️ **Требует права доступа к Docker socket**
   - Measurement: `docker`

5. **HTTP health check** (`[[inputs.http_response]]`):
   - Проверяет доступность `http://sso:8081/health`
   - Measurement: `http_response`

**Как работает:**

```
┌─────────────────────────────────────────────────────────────┐
│                    Telegraf Agent                           │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  INPUT PLUGINS (сбор данных)                         │  │
│  │                                                       │  │
│  │  [[inputs.prometheus]]                                │  │
│  │    urls = ["http://sso:8081/metrics"]                │  │
│  │    interval = 10s                                     │  │
│  │    ↓                                                  │  │
│  │    Скрейпит:                                          │  │
│  │    sso_http_requests_total{method="GET",...} 142      │  │
│  │    ↓                                                  │  │
│  │    Преобразует в Line Protocol:                       │  │
│  │    prometheus,__name__=sso_http_requests_total,       │  │
│  │      method=GET value=142 1732632120000000000         │  │
│  │                                                       │  │
│  │  [[inputs.postgresql]]                                │  │
│  │    connection_string = "postgres://..."              │  │
│  │    ↓                                                  │  │
│  │    postgresql,db=FourthCourseFirstProject             │  │
│  │      numbackends=5 1732632120000000000                │  │
│  │                                                       │  │
│  │  [[inputs.cpu]], [[inputs.mem]], [[inputs.disk]]     │  │
│  │    ↓                                                  │  │
│  │    cpu usage_user=15.2 1732632120000000000            │  │
│  └───────────────────────────────────────────────────────┘  │
│                          ↓                                  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  AGENT (буферизация и flush)                         │  │
│  │  flush_interval = 10s                                 │  │
│  │  metric_buffer_limit = 10000                          │  │
│  └───────────────────────────────────────────────────────┘  │
│                          ↓                                  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  OUTPUT PLUGIN                                        │  │
│  │                                                       │  │
│  │  [[outputs.influxdb_v2]]                              │  │
│  │    urls = ["http://influxdb:8086"]                    │  │
│  │    token = "my-super-secret-auth-token"               │  │
│  │    organization = "sso-org"                           │  │
│  │    bucket = "sso-metrics"                             │  │
│  │    ↓                                                  │  │
│  │    HTTP POST /api/v2/write?org=sso-org&bucket=...    │  │
│  │    Body (Line Protocol):                              │  │
│  │    prometheus,__name__=sso_http_requests_total,...    │  │
│  │    postgresql,db=FourthCourseFirstProject...          │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**Конфигурация:**

**Файл:** [`deployments/telegraf/telegraf.conf`](../deployments/telegraf/telegraf.conf)

```toml
[agent]
  interval = "10s"          # Как часто собирать метрики
  flush_interval = "10s"    # Как часто отправлять в InfluxDB

# OUTPUT: InfluxDB 2.x
[[outputs.influxdb_v2]]
  urls = ["http://influxdb:8086"]
  token = "my-super-secret-auth-token"
  organization = "sso-org"
  bucket = "sso-metrics"
  timeout = "5s"

# INPUT: Prometheus метрики из SSO
[[inputs.prometheus]]
  urls = ["http://sso:8081/metrics"]
  metric_version = 2
  response_timeout = "5s"
  [inputs.prometheus.tags]
    service = "sso-api"
    environment = "dev"

# INPUT: PostgreSQL
[[inputs.postgresql]]
  address = "host=host.docker.internal port=5432 user=phenirain password='phenirain13)' sslmode=disable dbname=FourthCourseFirstProject"
  [inputs.postgresql.tags]
    service = "main-postgres"
    environment = "dev"

# INPUT: System metrics
[[inputs.cpu]]
[[inputs.mem]]
[[inputs.disk]]
[[inputs.net]]

# INPUT: Docker (НЕ работает из-за permissions)
[[inputs.docker]]
  endpoint = "unix:///var/run/docker.sock"
  # ⚠️ Требует прав доступа к Docker socket
```

---

### 3. **InfluxDB 2.x**

**Роль:** Time Series Database для хранения метрик.

**Docker контейнер:** `sso-influxdb`
**Образ:** `influxdb:2.7-alpine`
**Порт:** `8086`

**Ключевые файлы:**
- [`deployments/docker-compose.yaml`](../deployments/docker-compose.yaml) - конфигурация контейнера
- [`deployments/grafana/provisioning/datasources/influxdb.yml`](../deployments/grafana/provisioning/datasources/influxdb.yml) - подключение к Grafana

**Структура данных:**

InfluxDB использует **Organization → Bucket → Measurement → Tags + Fields**.

```
Organization: sso-org
  ↓
Bucket: sso-metrics (retention 7 days)
  ↓
Measurements:
  - auth_operation
      Tags: operation, status, user_role, service, environment
      Fields: count, duration_ms

  - http_request
      Tags: method, path, status, service, environment
      Fields: duration_seconds, count

  - users
      Tags: service, environment, region
      Fields: total, active_last_24h, new_today

  - prometheus
      Tags: __name__, method, path, status, job, instance
      Fields: value (для counters/gauges), le (для histograms)

  - cpu, mem, disk, net, postgresql, docker
      Tags: зависит от input plugin
      Fields: зависит от метрики
```

**Line Protocol (формат данных):**

```
measurement,tag1=value1,tag2=value2 field1=value1,field2=value2 timestamp

# Пример:
auth_operation,operation=login,status=success,user_role=client,service=sso-api,environment=dev count=1,duration_ms=0 1732632892567153960

http_request,method=POST,path=/auth/logIn,status=200,service=sso-api,environment=dev duration_seconds=0.006,count=1 1732632892567153960

users,service=sso-api,environment=dev,region=local total=22,active_last_24h=0,new_today=0 1732632851395347719
```

**Инициализация:**

**Файл:** [`deployments/docker-compose.yaml`](../deployments/docker-compose.yaml)

```yaml
influxdb:
  image: influxdb:2.7-alpine
  environment:
    - DOCKER_INFLUXDB_INIT_MODE=setup
    - DOCKER_INFLUXDB_INIT_USERNAME=admin
    - DOCKER_INFLUXDB_INIT_PASSWORD=adminadmin
    - DOCKER_INFLUXDB_INIT_ORG=sso-org
    - DOCKER_INFLUXDB_INIT_BUCKET=sso-metrics
    - DOCKER_INFLUXDB_INIT_RETENTION=7d
    - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=my-super-secret-auth-token
  volumes:
    - influxdb-data:/var/lib/influxdb2
    - influxdb-config:/etc/influxdb2
```

**API Endpoints:**

- **Write API:** `POST /api/v2/write?org=sso-org&bucket=sso-metrics`
  - Принимает Line Protocol
  - Используется Telegraf и Go-клиентом

- **Query API:** `POST /api/v2/query?org=sso-org`
  - Принимает Flux queries
  - Используется Grafana

---

## Как работает InfluxDB

### Модель данных

InfluxDB хранит данные как **measurements** с **tags** (индексы) и **fields** (значения):

```
┌──────────────────────────────────────────────────────────┐
│  Measurement: auth_operation                             │
│                                                          │
│  Tags (indexed, searchable):                             │
│    operation = login                                     │
│    status = success                                      │
│    user_role = client   ← MEANINGFUL TAG!                │
│    service = sso-api                                     │
│    environment = dev                                     │
│                                                          │
│  Fields (values):                                        │
│    count = 1                                             │
│    duration_ms = 0                                       │
│                                                          │
│  Timestamp: 2025-11-26T15:34:52.567153960Z              │
└──────────────────────────────────────────────────────────┘
```

**Tags vs Fields:**

| Аспект | Tags | Fields |
|--------|------|--------|
| **Индексация** | ✅ Индексируются | ❌ Не индексируются |
| **Поиск** | Быстрый | Медленный |
| **Тип данных** | Только строки | Любые (int, float, string, bool) |
| **Кардинальность** | Должна быть низкой | Может быть любой |
| **Использование** | Фильтрация, группировка | Агрегация, математика |

**Пример:**
```flux
// Фильтрация по tag (быстро):
|> filter(fn: (r) => r["user_role"] == "admin")

// Фильтрация по field (медленно):
|> filter(fn: (r) => r["_value"] > 100)
```

### Осмысленный тег user_role

**Почему `user_role` - это хороший тег?**

1. **Низкая кардинальность:** Всего 4 значения (`client`, `manager`, `admin`, `unknown`)
2. **Бизнес-ценность:** Показывает активность разных типов пользователей
3. **Безопасность:** Помогает детектировать атаки на привилегированные аккаунты
4. **Масштабируемость:** Не растет с количеством пользователей

**Плохие примеры тегов:**
- ❌ `user_id` - высокая кардинальность (1000+ пользователей = 1000+ значений)
- ❌ `client_ip` - очень высокая кардинальность
- ❌ `request_id` - уникальное для каждого запроса

**Подробнее:** См. [`doc/USER_ROLE_TAG.md`](./USER_ROLE_TAG.md)

### Storage Engine (TSM)

InfluxDB использует **Time-Structured Merge Tree (TSM)** для хранения:

1. **Write-Ahead Log (WAL):**
   - Новые записи сначала идут в WAL (в памяти)
   - Гарантирует durability

2. **TSM Files:**
   - WAL периодически сбрасывается в TSM файлы на диске
   - Данные сжимаются и индексируются

3. **Retention Policy:**
   - Старые данные автоматически удаляются (у нас 7 дней)

---

## Роль Telegraf

### Зачем нужен Telegraf, если SSO пишет напрямую?

**SSO → InfluxDB (прямая запись):**
- ✅ **Низкая задержка** - метрики попадают мгновенно
- ✅ **Контроль над данными** - пишем именно то, что нужно
- ✅ **Осмысленные теги** - можем добавить `user_role`
- ❌ **Нет Prometheus метрик** - они остаются только в Prometheus

**Telegraf → InfluxDB (косвенная запись):**
- ✅ **Собирает Prometheus метрики** - дублирует их в InfluxDB
- ✅ **Дополнительные метрики** - PostgreSQL, Docker, CPU, memory
- ✅ **Единый источник** - все метрики в одном месте (InfluxDB)
- ❌ **Задержка 10 секунд** - не real-time

**Вывод:** Используем оба подхода!

```
SSO напрямую:
  auth_operation ──────────────────────────► InfluxDB
  http_request                              (мгновенно)
  users

Telegraf:
  Prometheus метрики ───► Преобразование ───► InfluxDB
  PostgreSQL метрики                         (каждые 10 сек)
  System метрики
```

### Преобразование Prometheus → InfluxDB

Telegraf преобразует Prometheus метрики в Line Protocol:

**Prometheus формат:**
```
sso_http_requests_total{method="GET",path="/health",status="200"} 142
```

**InfluxDB Line Protocol (после Telegraf):**
```
prometheus,__name__=sso_http_requests_total,method=GET,path=/health,status=200,service=sso-api,environment=dev value=142 1732632120000000000
```

**Flux запрос для получения:**
```flux
from(bucket: "sso-metrics")
  |> range(start: -10m)
  |> filter(fn: (r) => r["_measurement"] == "prometheus")
  |> filter(fn: (r) => r["__name__"] == "sso_http_requests_total")
  |> filter(fn: (r) => r["method"] == "GET")
```

---

## Прямая запись из Go

### Инициализация InfluxDB клиента

**Файл:** [`internal/application/server.go`](../internal/application/server.go)

```go
import (
    "github.com/phenirain/sso/pkg/metrics"
)

func NewServer(cfg *config.Config) (*Server, error) {
    // Инициализация метрик
    m := metrics.New()

    // Инициализация InfluxDB writer (если включен)
    if cfg.InfluxDB.Enabled {
        influxWriter, err := metrics.NewInfluxDBWriter(metrics.InfluxDBConfig{
            URL:    cfg.InfluxDB.URL,    // http://influxdb:8086
            Token:  cfg.InfluxDB.Token,  // my-super-secret-auth-token
            Org:    cfg.InfluxDB.Org,    // sso-org
            Bucket: cfg.InfluxDB.Bucket, // sso-metrics
        })
        if err != nil {
            log.Error("Failed to initialize InfluxDB writer", slog.String("error", err.Error()))
            // Продолжаем без InfluxDB - это опционально
        } else {
            m.InfluxDB = influxWriter
            log.Info("InfluxDB metrics writer initialized successfully")

            // Обработчик ошибок записи
            go func() {
                for err := range influxWriter.GetErrors() {
                    log.Error("InfluxDB write error", slog.String("error", err.Error()))
                }
            }()
        }
    }

    return &Server{
        metrics: m,
    }, nil
}
```

### Создание InfluxDB Writer

**Файл:** [`pkg/metrics/influxdb.go`](../pkg/metrics/influxdb.go)

```go
package metrics

import (
    "context"
    "time"
    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    "github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxDBWriter struct {
    client   influxdb2.Client
    writeAPI api.WriteAPI  // Асинхронный API
    org      string
    bucket   string
}

func NewInfluxDBWriter(cfg InfluxDBConfig) (*InfluxDBWriter, error) {
    // Создание клиента
    client := influxdb2.NewClient(cfg.URL, cfg.Token)

    // Проверка подключения
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    health, err := client.Health(ctx)
    if err != nil {
        return nil, fmt.Errorf("influxdb connection failed: %w", err)
    }

    if health.Status != "pass" {
        return nil, fmt.Errorf("influxdb health check failed: status=%s", health.Status)
    }

    // Асинхронный WriteAPI (буферизация + batch writes)
    writeAPI := client.WriteAPI(cfg.Org, cfg.Bucket)

    return &InfluxDBWriter{
        client:   client,
        writeAPI: writeAPI,
        org:      cfg.Org,
        bucket:   cfg.Bucket,
    }, nil
}
```

### Запись auth_operation метрики

**Файл:** [`pkg/metrics/influxdb.go`](../pkg/metrics/influxdb.go)

```go
func (w *InfluxDBWriter) WriteAuthOperation(operation, status, userRole, environment string, durationMs float64) {
    p := influxdb2.NewPoint(
        "auth_operation",  // measurement
        map[string]string{  // tags
            "operation":   operation,   // login, logout, refresh, register
            "status":      status,      // success, failure
            "user_role":   userRole,    // admin, manager, client, unknown
            "service":     "sso-api",
            "environment": environment, // dev, staging, prod
        },
        map[string]interface{}{  // fields
            "count":       1,
            "duration_ms": durationMs,
        },
        time.Now(),  // timestamp
    )
    w.writeAPI.WritePoint(p)  // Асинхронная запись (не блокирует)
}
```

### Использование в auth handler

**Файл:** [`internal/application/auth/handler.go`](../internal/application/auth/handler.go)

```go
func (h *Handler) LogIn(c echo.Context) error {
    // ... бизнес-логика ...

    result, err := h.s.Auth(ctx, req, isNew)
    if err != nil {
        // При ошибке роль неизвестна
        if h.m.InfluxDB != nil {
            h.m.InfluxDB.WriteAuthOperation("login", "failure", "unknown", "dev", 0)
        }
        return c.JSON(200, response.NewBadResponse[any]("Ошибка авторизации", err.Error()))
    }

    // При успехе используем реальную роль
    roleName := roleIDToName(result.RoleId)  // 1→client, 2→manager, 3→admin
    if h.m.InfluxDB != nil {
        h.m.InfluxDB.WriteAuthOperation("login", "success", roleName, "dev", 0)
    }

    return c.JSON(200, response.NewGoodResponse(result))
}
```

### Фоновое обновление users метрики

**Файл:** [`internal/application/server.go`](../internal/application/server.go)

```go
func (s *Server) startMetricsCollector() {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            // Получаем количество пользователей из БД
            total, err := s.userRepo.CountTotalUsers(context.Background())
            if err != nil {
                continue
            }

            // Обновляем Prometheus gauge
            s.metrics.SetTotalUsers(total)

            // Обновляем InfluxDB
            if s.metrics.InfluxDB != nil {
                s.metrics.InfluxDB.WriteTotalUsers(total, 0, 0, "dev", "local")
            }
        }
    }()
}
```

---

## Визуализация в Grafana

### Подключение InfluxDB Datasource

**Файл:** [`deployments/grafana/provisioning/datasources/influxdb.yml`](../deployments/grafana/provisioning/datasources/influxdb.yml)

```yaml
apiVersion: 1

datasources:
  - name: InfluxDB
    type: influxdb
    uid: P951FEA4DE68E13C5
    access: proxy
    url: http://influxdb:8086
    jsonData:
      version: Flux           # InfluxDB 2.x использует Flux query language
      organization: sso-org
      defaultBucket: sso-metrics
      tlsSkipVerify: true
    secureJsonData:
      token: my-super-secret-auth-token
    editable: true
```

### Flux Query Language

**Базовая структура:**

```flux
from(bucket: "sso-metrics")                      // Источник данных
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)  // Временной диапазон
  |> filter(fn: (r) => r["_measurement"] == "auth_operation")  // Фильтр по measurement
  |> filter(fn: (r) => r["_field"] == "count")   // Фильтр по field
  |> filter(fn: (r) => r["user_role"] == "client")  // Фильтр по tag
  |> sum()                                       // Агрегация
```

**Переменные Grafana:**
- `v.timeRangeStart` - начало временного диапазона из UI
- `v.timeRangeStop` - конец временного диапазона
- `v.windowPeriod` - период агрегации (auto, 1m, 5m, ...)

### Примеры панелей

#### 1. HTTP Requests per Second

**Файл:** [`deployments/grafana/provisioning/dashboards/dashboards/influxdb-sso-metrics.json`](../deployments/grafana/provisioning/dashboards/dashboards/influxdb-sso-metrics.json)

```flux
from(bucket: "sso-metrics")
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)
  |> filter(fn: (r) => r["_measurement"] == "http_request")
  |> filter(fn: (r) => r["_field"] == "count")
  |> filter(fn: (r) => r["service"] == "sso-api")
  |> group(columns: ["method", "path", "status"])
  |> aggregateWindow(every: v.windowPeriod, fn: sum, createEmpty: false)
  |> map(fn: (r) => ({ r with _value: float(v: r._value) / float(v: uint(v: v.windowPeriod)) * 1000000000.0 }))
```

**Что делает:**
1. Берет measurement `http_request`
2. Фильтрует по field `count`
3. Группирует по method, path, status
4. Суммирует за каждое окно (`v.windowPeriod`)
5. Преобразует в requests/sec (делит на длительность окна в наносекундах)

#### 2. Client Auth Success (Gauge)

```flux
from(bucket: "sso-metrics")
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)
  |> filter(fn: (r) => r["_measurement"] == "auth_operation")
  |> filter(fn: (r) => r["_field"] == "count")
  |> filter(fn: (r) => r["service"] == "sso-api")
  |> filter(fn: (r) => r["status"] == "success")
  |> filter(fn: (r) => r["user_role"] == "client")  ← Фильтр по роли!
  |> sum()
  |> yield(name: "client_success")
```

**Что показывает:**
- Общее количество успешных авторизаций клиентов за выбранный период

**Аналогично для:**
- Manager Auth Success: `user_role == "manager"`
- Admin Auth Success: `user_role == "admin"`
- Client Auth Failure: `status == "failure" AND user_role == "client"`
- И т.д.

#### 3. HTTP Duration p95

```flux
from(bucket: "sso-metrics")
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)
  |> filter(fn: (r) => r["_measurement"] == "http_request")
  |> filter(fn: (r) => r["_field"] == "duration_seconds")
  |> filter(fn: (r) => r["service"] == "sso-api")
  |> group(columns: ["method", "path"])
  |> aggregateWindow(every: v.windowPeriod, fn: (column, tables=<-) => tables |> quantile(q: 0.95, column: column), createEmpty: false)
```

**Что делает:**
1. Берет `duration_seconds` field
2. Группирует по method, path
3. Вычисляет 95-й перцентиль за каждое окно

---

## Конфигурационные файлы

### Docker Compose (полная версия)

**Файл:** [`deployments/docker-compose.yaml`](../deployments/docker-compose.yaml)

```yaml
services:
  sso:
    image: phenirain/fourthcoursefirstproject-sso:latest
    ports:
      - "8081:8081"
    depends_on:
      - influxdb
    environment:
      # Конфиг InfluxDB берется из config.yaml

  influxdb:
    image: influxdb:2.7-alpine
    container_name: sso-influxdb
    environment:
      - DOCKER_INFLUXDB_INIT_MODE=setup
      - DOCKER_INFLUXDB_INIT_USERNAME=admin
      - DOCKER_INFLUXDB_INIT_PASSWORD=adminadmin
      - DOCKER_INFLUXDB_INIT_ORG=sso-org
      - DOCKER_INFLUXDB_INIT_BUCKET=sso-metrics
      - DOCKER_INFLUXDB_INIT_RETENTION=7d
      - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=my-super-secret-auth-token
    ports:
      - "8086:8086"
    volumes:
      - influxdb-data:/var/lib/influxdb2
      - influxdb-config:/etc/influxdb2

  telegraf:
    image: telegraf:1.31-alpine
    container_name: sso-telegraf
    volumes:
      - ./telegraf/telegraf.conf:/etc/telegraf/telegraf.conf:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    depends_on:
      - influxdb

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/provisioning/datasources:/etc/grafana/provisioning/datasources
      - ./grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards
    depends_on:
      - influxdb

volumes:
  influxdb-data:
  influxdb-config:
  grafana-data:
```

### Telegraf Configuration (полная версия)

**Файл:** [`deployments/telegraf/telegraf.conf`](../deployments/telegraf/telegraf.conf)

```toml
[agent]
  interval = "10s"
  round_interval = true
  metric_batch_size = 1000
  metric_buffer_limit = 10000
  collection_jitter = "0s"
  flush_interval = "10s"
  flush_jitter = "0s"
  precision = "0s"
  hostname = "sso-telegraf"
  omit_hostname = false

###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################

[[outputs.influxdb_v2]]
  urls = ["http://influxdb:8086"]
  token = "my-super-secret-auth-token"
  organization = "sso-org"
  bucket = "sso-metrics"
  timeout = "5s"

###############################################################################
#                            INPUT PLUGINS                                    #
###############################################################################

# PostgreSQL Database Monitoring
[[inputs.postgresql]]
  address = "host=host.docker.internal port=5432 user=phenirain password='phenirain13)' sslmode=disable dbname=FourthCourseFirstProject"
  [inputs.postgresql.tags]
    service = "main-postgres"
    environment = "dev"

# System Metrics
[[inputs.cpu]]
  percpu = true
  totalcpu = true
  [inputs.cpu.tags]
    service = "sso-system"

[[inputs.mem]]
  [inputs.mem.tags]
    service = "sso-system"

[[inputs.disk]]
  ignore_fs = ["tmpfs", "devtmpfs", "devfs"]
  [inputs.disk.tags]
    service = "sso-system"

[[inputs.net]]
  interfaces = ["eth*"]
  [inputs.net.tags]
    service = "sso-system"

# Docker Container Metrics (требует permissions)
[[inputs.docker]]
  endpoint = "unix:///var/run/docker.sock"
  container_name_include = ["sso*", "postgres*"]
  timeout = "5s"
  [inputs.docker.tags]
    service = "sso-docker"

# HTTP Health Check
[[inputs.http_response]]
  urls = ["http://sso:8081/health"]
  response_timeout = "5s"
  method = "GET"
  [inputs.http_response.tags]
    service = "sso-api"
    check_type = "health"

# Prometheus Metrics Scraper (КЛЮЧЕВОЙ INPUT!)
[[inputs.prometheus]]
  urls = ["http://sso:8081/metrics"]
  metric_version = 2
  response_timeout = "5s"
  [inputs.prometheus.tags]
    service = "sso-api"
    environment = "dev"
```

---

## Диаграмма взаимодействия

```
┌──────────────────────────────────────────────────────────────────┐
│                        Временная шкала                           │
└──────────────────────────────────────────────────────────────────┘

T=0s    SSO получает HTTP запрос POST /auth/logIn от клиента
         ↓
        SSO обрабатывает запрос (6ms)
         ↓
        SSO определяет роль пользователя: roleIDToName(1) = "client"
         ↓
        SSO вызывает:
          1) metrics.RecordRequest() → Prometheus counter увеличивается
          2) influxDB.WriteAuthOperation("login", "success", "client", "dev", 0)
         ↓
        InfluxDB WriteAPI (асинхронно):
          - Буферизирует точку данных в памяти
          - Батчит с другими точками
          - Отправляет HTTP POST /api/v2/write через ~1-2 секунды

T=1-2s  InfluxDB получает batch от Go-клиента:
        POST /api/v2/write?org=sso-org&bucket=sso-metrics
        Body: auth_operation,operation=login,status=success,user_role=client,...
         ↓
        InfluxDB записывает в WAL (Write-Ahead Log)
         ↓
        Данные доступны для запросов НЕМЕДЛЕННО

T=10s   Telegraf выполняет scrape:
        1) GET http://sso:8081/metrics
           ↓
           SSO отвечает Prometheus метриками
           ↓
           Telegraf преобразует в Line Protocol:
           prometheus,__name__=sso_http_requests_total,... value=8

        2) Подключается к PostgreSQL
           ↓
           SELECT * FROM pg_stat_database
           ↓
           postgresql,db=FourthCourseFirstProject numbackends=5

        3) Читает /proc/stat, /proc/meminfo
           ↓
           cpu usage_user=15.2
           mem used_percent=45.3
         ↓
        Telegraf отправляет batch в InfluxDB:
        POST /api/v2/write
        Body: все собранные метрики в Line Protocol

T=10s   Grafana выполняет Flux запрос (если дашборд открыт):
        POST /api/v2/query?org=sso-org
        Body: from(bucket: "sso-metrics") |> range(...) |> filter(...)
         ↓
        InfluxDB выполняет запрос:
          - Читает данные из WAL + TSM files
          - Применяет фильтры и агрегации
          - Возвращает результат в CSV формате
         ↓
        Grafana парсит CSV и рисует график

T=15s   Grafana обновляет дашборд (refresh: "5s")
        → Новый Flux запрос
        → Новые данные на графике

T=20s   Telegraf делает следующий scrape...
        Цикл повторяется
```

---

## Проверка работы

### 1. Проверить подключение SSO к InfluxDB

```bash
docker compose -f deployments/docker-compose.yaml logs sso | grep -i influx
```

**Ожидаемый вывод:**
```
{"level":"INFO","msg":"InfluxDB metrics writer initialized successfully"}
```

❌ **Если ошибка:**
```
{"level":"ERROR","msg":"Failed to initialize InfluxDB writer","error":"..."}
```
→ Перезапустить SSO: `docker compose -f deployments/docker-compose.yaml restart sso`

### 2. Проверить данные в InfluxDB

```bash
# Проверить measurements
docker exec sso-influxdb influx query \
  --org sso-org \
  --token my-super-secret-auth-token \
  'import "influxdata/influxdb/schema"
   schema.measurements(bucket: "sso-metrics")'
```

**Ожидаемый вывод:**
```
auth_operation
http_request
users
prometheus
cpu
mem
disk
...
```

### 3. Проверить auth_operation данные

```bash
docker exec sso-influxdb influx query \
  --org sso-org \
  --token my-super-secret-auth-token \
  'from(bucket: "sso-metrics")
   |> range(start: -10m)
   |> filter(fn: (r) => r["_measurement"] == "auth_operation")
   |> limit(n: 5)'
```

**Ожидаемый вывод:**
```
_time                 _field       _value  operation  status   user_role
2025-11-26T15:34:52Z  count        1       login      success  client
2025-11-26T15:34:52Z  duration_ms  0       login      success  client
```

### 4. Сгенерировать тестовые метрики

```bash
# Неудачная авторизация (user_role=unknown)
curl -X POST http://localhost:8081/auth/logIn \
  -H "Content-Type: application/json" \
  -d '{"login":"invalid@test.com","password":"wrongpass"}'

# Подождать 12 секунд (flush_interval + немного)
sleep 12

# Проверить данные
docker exec sso-influxdb influx query \
  --org sso-org \
  --token my-super-secret-auth-token \
  'from(bucket: "sso-metrics")
   |> range(start: -2m)
   |> filter(fn: (r) => r["_measurement"] == "auth_operation")
   |> filter(fn: (r) => r["user_role"] == "unknown")'
```

### 5. Открыть Grafana дашборд

```bash
# Открыть в браузере
open http://localhost:3000
# Логин: admin / admin

# Перейти в Dashboards → SSO API Metrics Dashboard (InfluxDB)
```

**Ожидается:**
- Gauge метрики для Client/Manager/Admin Auth Success/Failure
- Графики HTTP Requests, Duration, Total Users

---

## Troubleshooting

### SSO не подключается к InfluxDB

**Проблема:**
```
Failed to initialize InfluxDB writer: connection refused
```

**Решение:**
```bash
# 1. Проверить, что InfluxDB запущен
docker ps | grep influxdb

# 2. Проверить health
docker exec sso-influxdb influx ping

# 3. Перезапустить SSO (чтобы переподключиться)
docker compose -f deployments/docker-compose.yaml restart sso
```

### Telegraf не может скрейпить Prometheus метрики

**Проблема:**
```bash
docker compose -f deployments/docker-compose.yaml logs telegraf
# Нет ошибок, но данные не попадают
```

**Решение:**
```bash
# Проверить доступность /metrics из Telegraf
docker exec sso-telegraf wget -O- http://sso:8081/metrics

# Проверить конфигурацию
docker exec sso-telegraf cat /etc/telegraf/telegraf.conf | grep -A5 "inputs.prometheus"
```

### Docker metrics не собираются (permission denied)

**Проблема:**
```
Error in plugin: permission denied while trying to connect to the Docker daemon socket
```

**Это нормально!** Docker socket требует специальных прав. Можно:
1. Игнорировать (остальные метрики работают)
2. Добавить Telegraf в группу docker:
```yaml
telegraf:
  user: "telegraf:999"  # 999 = docker group ID
```

### Grafana показывает "No data"

**Проблема:** Панели пустые

**Решение:**
1. **Проверить time range** - должен быть "Last 15 minutes" или больше
2. **Проверить datasource** - должен быть InfluxDB, а не Prometheus
3. **Проверить данные в InfluxDB** (см. выше)
4. **Сгенерировать тестовые метрики**
5. **Проверить Flux запрос** - скопировать в InfluxDB UI и выполнить

---

## Полезные ссылки

- **InfluxDB 2.x Docs:** https://docs.influxdata.com/influxdb/v2/
- **Flux Language:** https://docs.influxdata.com/flux/v0/
- **Telegraf Plugins:** https://docs.influxdata.com/telegraf/v1/plugins/
- **InfluxDB Go Client:** https://github.com/influxdata/influxdb-client-go
- **Line Protocol:** https://docs.influxdata.com/influxdb/v2/reference/syntax/line-protocol/

---

**Дата создания:** 2025-11-26
**Версия:** 1.0
