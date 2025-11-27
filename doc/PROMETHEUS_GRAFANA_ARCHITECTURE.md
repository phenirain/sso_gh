# Prometheus + Grafana: Архитектура и взаимодействие

## 📋 Содержание
1. [Общая архитектура](#общая-архитектура)
2. [Компоненты системы](#компоненты-системы)
3. [Как работает Prometheus](#как-работает-prometheus)
4. [Интеграция с Go приложением](#интеграция-с-go-приложением)
5. [Визуализация в Grafana](#визуализация-в-grafana)
6. [Конфигурационные файлы](#конфигурационные-файлы)

---

## Общая архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                    SSO Service (Go)                         │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Prometheus Metrics Exporter (/metrics endpoint)    │   │
│  │  - HTTP request counters                            │   │
│  │  - HTTP request duration histograms                 │   │
│  │  - Auth operations counters                         │   │
│  │  - Total users gauge                                │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                          ↓ HTTP GET /metrics (каждые 5 сек)
┌─────────────────────────────────────────────────────────────┐
│                    Prometheus Server                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Time Series Database (TSDB)                        │   │
│  │  - Хранит метрики за последние 7 дней              │   │
│  │  - Индексирует по labels (method, path, status)    │   │
│  │  - Предоставляет PromQL query API                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                          ↓ PromQL запросы (каждые 5 сек)
┌─────────────────────────────────────────────────────────────┐
│                         Grafana                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Dashboard: SSO API Metrics Dashboard               │   │
│  │  - HTTP Requests per Second (irate)                │   │
│  │  - HTTP Duration Percentiles (histogram_quantile)  │   │
│  │  - Average Response Time                           │   │
│  │  - Total Registered Users                          │   │
│  │  - Auth Success/Failure Counts                     │   │
│  │  - Total HTTP Requests                             │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                          ↓ Пользователь видит в браузере
                    http://localhost:3000
```

---

## Компоненты системы

### 1. **SSO Service (Go приложение)**

**Роль:** Генерирует метрики и экспортирует их в формате Prometheus.

**Ключевые файлы:**
- [`pkg/metrics/metrics.go`](../pkg/metrics/metrics.go) - основной модуль метрик
- [`internal/application/server.go`](../internal/application/server.go) - инициализация метрик
- [`internal/application/auth/handler.go`](../internal/application/auth/handler.go) - запись auth метрик

**Метрики, которые экспортируются:**
- `sso_http_requests_total{method, path, status}` - счётчик HTTP запросов
- `sso_http_request_duration_seconds{method, path}` - гистограмма длительности запросов
- `sso_auth_operations_total{operation, status}` - счётчик auth операций
- `sso_total_users` - gauge общего количества пользователей

**Endpoint:** `http://sso:8081/metrics`

**Формат данных (Prometheus text format):**
```prometheus
# HELP sso_http_requests_total Total number of HTTP requests
# TYPE sso_http_requests_total counter
sso_http_requests_total{method="GET",path="/health",status="200"} 142

# HELP sso_http_request_duration_seconds HTTP request duration
# TYPE sso_http_request_duration_seconds histogram
sso_http_request_duration_seconds_bucket{method="POST",path="/auth/logIn",le="0.005"} 5
sso_http_request_duration_seconds_bucket{method="POST",path="/auth/logIn",le="0.01"} 8
sso_http_request_duration_seconds_sum{method="POST",path="/auth/logIn"} 0.042
sso_http_request_duration_seconds_count{method="POST",path="/auth/logIn"} 8

# HELP sso_total_users Total number of registered users
# TYPE sso_total_users gauge
sso_total_users 22
```

---

### 2. **Prometheus Server**

**Роль:** Скрейпит (собирает) метрики из SSO сервиса, хранит в TSDB и предоставляет PromQL API.

**Docker контейнер:** `sso-prometheus`
**Образ:** `prom/prometheus:latest`
**Порт:** `9090`

**Ключевые файлы:**
- [`deployments/prometheus/prometheus.yml`](../deployments/prometheus/prometheus.yml) - конфигурация Prometheus
- [`deployments/docker-compose.yaml`](../deployments/docker-compose.yaml) - запуск контейнера

**Как работает:**

1. **Scraping (сбор метрик):**
   - Каждые 5 секунд (`scrape_interval: 5s`) Prometheus делает HTTP GET запрос к `http://sso:8081/metrics`
   - Парсит текстовый формат Prometheus
   - Сохраняет данные в Time Series Database (TSDB)

2. **Хранение:**
   - Данные хранятся в volume `prometheus-data:/prometheus`
   - Retention: 7 дней (`--storage.tsdb.retention.time=7d`)
   - После 7 дней старые данные автоматически удаляются

3. **Индексация:**
   - Метрики индексируются по **labels** (method, path, status)
   - Это позволяет делать быстрые запросы типа: "все запросы с status=500"

4. **Query API:**
   - Предоставляет HTTP API для выполнения PromQL запросов
   - Endpoint: `http://prometheus:9090/api/v1/query`

**Пример PromQL запроса:**
```promql
# Requests per second (rate of change)
irate(sso_http_requests_total[30s])

# 95th percentile request duration
histogram_quantile(0.95, sum(rate(sso_http_request_duration_seconds_bucket[5m])) by (le, method, path))

# Average response time
sum(rate(sso_http_request_duration_seconds_sum[5m])) / sum(rate(sso_http_request_duration_seconds_count[5m]))
```

---

### 3. **Grafana**

**Роль:** Визуализирует метрики из Prometheus через дашборды.

**Docker контейнер:** `sso-grafana`
**Образ:** `grafana/grafana:latest`
**Порт:** `3000`
**Логин/пароль:** `admin/admin`

**Ключевые файлы:**
- [`deployments/grafana/provisioning/datasources/prometheus.yml`](../deployments/grafana/provisioning/datasources/prometheus.yml) - подключение к Prometheus
- [`deployments/grafana/provisioning/dashboards/dashboards.yml`](../deployments/grafana/provisioning/dashboards/dashboards.yml) - автоматическая загрузка дашбордов
- [`deployments/grafana/provisioning/dashboards/dashboards/sso-metrics.json`](../deployments/grafana/provisioning/dashboards/dashboards/sso-metrics.json) - дашборд с панелями

**Как работает:**

1. **Datasource (источник данных):**
   - Grafana подключается к Prometheus через HTTP API: `http://prometheus:9090`
   - Использует PromQL для запросов данных
   - Конфигурация в [`prometheus.yml`](../deployments/grafana/provisioning/datasources/prometheus.yml)

2. **Dashboard (дашборд):**
   - Состоит из **panels** (панелей) - отдельных графиков/таблиц
   - Каждая панель выполняет PromQL запрос к Prometheus
   - Обновляется каждые 5 секунд (`refresh: "5s"`)

3. **Query Flow (поток запросов):**
   ```
   Grafana Panel → PromQL запрос → Prometheus API → TSDB → результат → визуализация
   ```

**Пример панели в JSON:**
```json
{
  "title": "HTTP Requests per Second",
  "targets": [
    {
      "expr": "irate(sso_http_requests_total[30s])",
      "legendFormat": "{{method}} {{path}} - {{status}}"
    }
  ],
  "type": "timeseries"
}
```

---

## Как работает Prometheus

### Модель данных

Prometheus хранит данные как **time series** - последовательности значений с временными метками:

```
metric_name{label1="value1", label2="value2"} value timestamp
```

**Пример:**
```
sso_http_requests_total{method="GET", path="/health", status="200"} 142 1732632120
sso_http_requests_total{method="POST", path="/auth/logIn", status="200"} 8 1732632120
```

### Типы метрик

1. **Counter (счётчик)** - только растёт, никогда не уменьшается:
   - `sso_http_requests_total` - всего запросов с момента старта
   - `sso_auth_operations_total` - всего auth операций

2. **Gauge (датчик)** - может расти и уменьшаться:
   - `sso_total_users` - текущее количество пользователей

3. **Histogram (гистограмма)** - распределение значений по buckets:
   - `sso_http_request_duration_seconds_bucket{le="0.005"}` - запросы до 5ms
   - `sso_http_request_duration_seconds_bucket{le="0.01"}` - запросы до 10ms
   - `sso_http_request_duration_seconds_sum` - сумма всех длительностей
   - `sso_http_request_duration_seconds_count` - количество запросов

### Scraping Process (процесс сбора)

1. **Prometheus делает HTTP GET запрос:**
   ```
   GET http://sso:8081/metrics
   ```

2. **SSO отвечает текстом в формате Prometheus:**
   ```
   sso_http_requests_total{method="GET",path="/health",status="200"} 142
   ```

3. **Prometheus парсит и сохраняет в TSDB:**
   - Создаёт time series: `sso_http_requests_total{method="GET",path="/health",status="200"}`
   - Добавляет новую точку: `(timestamp, 142)`

4. **Повторяет каждые 5 секунд**

---

## Интеграция с Go приложением

### Шаг 1: Подключение библиотеки

**Файл:** [`go.mod`](../go.mod)

```go
require (
    github.com/prometheus/client_golang v1.20.5
)
```

### Шаг 2: Создание метрик

**Файл:** [`pkg/metrics/metrics.go`](../pkg/metrics/metrics.go)

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Counter - HTTP requests
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "sso_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"}, // labels
    )

    // Histogram - request duration
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "sso_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
        },
        []string{"method", "path"},
    )

    // Gauge - total users
    totalUsers = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "sso_total_users",
            Help: "Total number of registered users",
        },
    )
)

// RecordRequest записывает HTTP запрос
func (m *Metrics) RecordRequest(method, path, status string, duration float64) {
    httpRequestsTotal.WithLabelValues(method, path, status).Inc()
    httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// SetTotalUsers обновляет количество пользователей
func (m *Metrics) SetTotalUsers(count int) {
    totalUsers.Set(float64(count))
}
```

### Шаг 3: HTTP Endpoint для метрик

**Файл:** [`internal/application/server.go`](../internal/application/server.go)

```go
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewServer(cfg *config.Config, m *metrics.Metrics) *Server {
    e := echo.New()

    // Prometheus metrics endpoint
    e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

    return &Server{echo: e}
}
```

### Шаг 4: Запись метрик в middleware

**Файл:** [`internal/application/auth/handler.go`](../internal/application/auth/handler.go)

```go
func (h *Handler) LogIn(c echo.Context) error {
    start := time.Now()

    // ... бизнес-логика ...

    // Записываем метрику
    duration := time.Since(start).Seconds()
    h.m.RecordRequest("POST", "/auth/logIn", "200", duration)
    h.m.RecordAuthOperation("login", "success")

    return c.JSON(200, response)
}
```

### Шаг 5: Фоновое обновление gauge метрик

**Файл:** [`internal/application/server.go`](../internal/application/server.go)

```go
// Обновляем количество пользователей каждые 30 секунд
func (s *Server) startMetricsCollector() {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            count, err := s.userRepo.CountTotalUsers(context.Background())
            if err == nil {
                s.metrics.SetTotalUsers(count)
            }
        }
    }()
}
```

---

## Визуализация в Grafana

### Подключение Prometheus Datasource

**Файл:** [`deployments/grafana/provisioning/datasources/prometheus.yml`](../deployments/grafana/provisioning/datasources/prometheus.yml)

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    uid: PBFA97CFB590B2093  # Уникальный ID
    access: proxy
    url: http://prometheus:9090  # URL Prometheus внутри Docker сети
    isDefault: true
    editable: true
```

**Что это делает:**
- Grafana автоматически подключается к Prometheus при старте
- `access: proxy` - Grafana сам делает запросы к Prometheus (не браузер пользователя)
- `isDefault: true` - этот datasource используется по умолчанию

### Автозагрузка дашбордов

**Файл:** [`deployments/grafana/provisioning/dashboards/dashboards.yml`](../deployments/grafana/provisioning/dashboards/dashboards.yml)

```yaml
apiVersion: 1

providers:
  - name: 'SSO Dashboards'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards/dashboards
```

**Что это делает:**
- Grafana сканирует папку `/etc/grafana/provisioning/dashboards/dashboards`
- Автоматически импортирует все `.json` файлы как дашборды
- `updateIntervalSeconds: 10` - проверяет изменения каждые 10 секунд
- `allowUiUpdates: true` - можно редактировать через UI

### Структура дашборда

**Файл:** [`deployments/grafana/provisioning/dashboards/dashboards/sso-metrics.json`](../deployments/grafana/provisioning/dashboards/dashboards/sso-metrics.json)

```json
{
  "title": "SSO API Metrics Dashboard",
  "uid": "sso-metrics",
  "refresh": "5s",  // Обновление каждые 5 секунд
  "panels": [
    {
      "id": 1,
      "title": "HTTP Requests per Second",
      "type": "timeseries",
      "targets": [
        {
          "expr": "irate(sso_http_requests_total[30s])",
          "legendFormat": "{{method}} {{path}} - {{status}}"
        }
      ]
    }
  ]
}
```

**Панели в дашборде:**

1. **HTTP Requests per Second**
   - PromQL: `irate(sso_http_requests_total[30s])`
   - Показывает мгновенную скорость прироста счётчика (requests/sec)

2. **HTTP Request Duration (Percentiles)**
   - PromQL: `histogram_quantile(0.95, sum(rate(sso_http_request_duration_seconds_bucket[5m])) by (le, method, path))`
   - Показывает 95-й перцентиль времени ответа

3. **Average Response Time**
   - PromQL: `sum(rate(sso_http_request_duration_seconds_sum[5m])) / sum(rate(sso_http_request_duration_seconds_count[5m]))`
   - Среднее время ответа

4. **Total Registered Users**
   - PromQL: `sso_total_users`
   - Текущее количество пользователей

5. **Auth Success Count**
   - PromQL: `sum(sso_auth_operations_total{status="success"})`
   - Количество успешных авторизаций

6. **Auth Failure Count**
   - PromQL: `sum(sso_auth_operations_total{status="failure"})`
   - Количество неудачных авторизаций

---

## Конфигурационные файлы

### Docker Compose

**Файл:** [`deployments/docker-compose.yaml`](../deployments/docker-compose.yaml)

```yaml
services:
  # SSO Service
  sso:
    image: phenirain/fourthcoursefirstproject-sso:latest
    ports:
      - "8081:8081"  # HTTP API + /metrics endpoint

  # Prometheus
  prometheus:
    image: prom/prometheus:latest
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=7d'
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    depends_on:
      - sso

  # Grafana
  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/provisioning/datasources:/etc/grafana/provisioning/datasources
      - ./grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards
    depends_on:
      - prometheus

volumes:
  prometheus-data:
  grafana-data:
```

### Prometheus Configuration

**Файл:** [`deployments/prometheus/prometheus.yml`](../deployments/prometheus/prometheus.yml)

```yaml
global:
  scrape_interval: 5s       # Как часто скрейпить метрики
  evaluation_interval: 5s   # Как часто проверять правила

scrape_configs:
  - job_name: 'sso-api'
    static_configs:
      - targets: ['sso:8081']  # Адрес SSO сервиса
    metrics_path: '/metrics'   # Путь к endpoint
```

**Параметры:**
- `scrape_interval: 5s` - каждые 5 секунд Prometheus делает GET `/metrics`
- `job_name: 'sso-api'` - имя job (появится как label `job="sso-api"` в метриках)
- `targets: ['sso:8081']` - список эндпоинтов для скрейпинга

---

## Диаграмма взаимодействия

```
┌──────────────────────────────────────────────────────────────────┐
│                        Временная шкала                           │
└──────────────────────────────────────────────────────────────────┘

T=0s    SSO получает HTTP запрос POST /auth/logIn
         ↓
        SSO обрабатывает запрос (6ms)
         ↓
        SSO вызывает metrics.RecordRequest("POST", "/auth/logIn", "200", 0.006)
         ↓
        Prometheus Counter увеличивается: sso_http_requests_total{...} 7 → 8
        Prometheus Histogram записывает: sso_http_request_duration_seconds_bucket{le="0.01"} 7 → 8

T=5s    Prometheus делает GET http://sso:8081/metrics
         ↓
        SSO отвечает текстом:
        sso_http_requests_total{method="POST",path="/auth/logIn",status="200"} 8
         ↓
        Prometheus сохраняет в TSDB:
        sso_http_requests_total{method="POST",path="/auth/logIn",status="200"} = (1732632125, 8)

T=5s    Grafana выполняет PromQL запрос:
        irate(sso_http_requests_total{path="/auth/logIn"}[30s])
         ↓
        Prometheus вычисляет:
        (8 - 7) / (5s - 0s) = 0.2 requests/sec
         ↓
        Grafana рисует точку на графике: (15:35:05, 0.2)

T=10s   Процесс повторяется...
```

---

## Проверка работы

### 1. Проверить метрики в SSO
```bash
curl http://localhost:8081/metrics | grep sso_
```

**Ожидаемый вывод:**
```
sso_http_requests_total{method="GET",path="/health",status="200"} 142
sso_total_users 22
```

### 2. Проверить Prometheus targets
Открыть: http://localhost:9090/targets

**Ожидается:**
- Job: `sso-api`
- State: `UP`
- Last Scrape: несколько секунд назад

### 3. Проверить PromQL запрос
Открыть: http://localhost:9090/graph

Ввести запрос:
```promql
sso_http_requests_total
```

**Ожидается:** таблица со всеми метриками HTTP запросов

### 4. Проверить Grafana datasource
Открыть: http://localhost:3000/connections/datasources

**Ожидается:**
- Prometheus datasource с зелёным статусом

### 5. Открыть дашборд
Открыть: http://localhost:3000/dashboards

**Ожидается:**
- "SSO API Metrics Dashboard" с графиками

---

## Troubleshooting

### Prometheus не скрейпит метрики

**Проблема:** `State: DOWN` в targets

**Решение:**
```bash
# Проверить доступность /metrics из контейнера Prometheus
docker exec sso-prometheus wget -O- http://sso:8081/metrics

# Проверить логи Prometheus
docker compose -f deployments/docker-compose.yaml logs prometheus
```

### Grafana не видит Prometheus

**Проблема:** `Error reading Prometheus` в datasource

**Решение:**
```bash
# Проверить доступность Prometheus из контейнера Grafana
docker exec sso-grafana wget -O- http://prometheus:9090/api/v1/query?query=up

# Проверить provisioning
docker exec sso-grafana ls -la /etc/grafana/provisioning/datasources/
```

### Дашборд не показывает данные

**Проблема:** "No data" на панелях

**Решение:**
1. Проверить time range в Grafana (должен быть "Last 15 minutes")
2. Проверить PromQL запрос в панели
3. Сгенерировать тестовые метрики:
```bash
curl -X POST http://localhost:8081/auth/logIn \
  -H "Content-Type: application/json" \
  -d '{"login":"test@test.com","password":"test"}'
```

---

## Полезные ссылки

- **Prometheus Documentation:** https://prometheus.io/docs/
- **PromQL Basics:** https://prometheus.io/docs/prometheus/latest/querying/basics/
- **Grafana Provisioning:** https://grafana.com/docs/grafana/latest/administration/provisioning/
- **Go Prometheus Client:** https://github.com/prometheus/client_golang

---

**Дата создания:** 2025-11-26
**Версия:** 1.0
