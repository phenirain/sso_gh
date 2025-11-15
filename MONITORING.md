# Мониторинг SSO API с помощью Prometheus и Grafana

## Обзор

Этот проект включает полную систему мониторинга на основе Prometheus и Grafana для отслеживания производительности и работоспособности SSO API Gateway.

## Метрики

### Стандартные HTTP метрики

1. **HTTP Request Rate** (`sso_http_requests_total`)
   - Тип: Counter
   - Описание: Общее количество HTTP запросов
   - Метки: `method`, `path`, `status`
   - Визуализация: Time series (линейный график с областью)

2. **HTTP Request Duration** (`sso_http_request_duration_seconds`)
   - Тип: Histogram
   - Описание: Длительность обработки HTTP запросов в секундах
   - Метки: `method`, `path`
   - Визуализация: Time series с перцентилями (p50, p95)

3. **HTTP Requests In-Flight** (`sso_http_requests_in_flight`)
   - Тип: Gauge
   - Описание: Текущее количество обрабатываемых HTTP запросов
   - Визуализация: Gauge (шкала)

### Кастомные бизнес-метрики

#### 1. Total Users (`sso_total_users`)
- **Тип**: Gauge
- **Описание**: Общее количество зарегистрированных пользователей в системе
- **Сохранность**: ✅ Данные сохраняются при перезапуске (читаются из БД)
- **Обновление**: Каждые 30 секунд
- **Визуализация**: Gauge (шкала)
- **Бизнес-значение**: Показывает рост пользовательской базы, критично для анализа масштабирования

#### 2. Authentication Operations (`sso_auth_operations_total`)
- **Тип**: Counter
- **Описание**: Общее количество операций аутентификации
- **Метки**:
  - `operation`: login, signup, refresh
  - `status`: success, failure
- **Сохранность**: ⚠️ Счетчик сбрасывается при перезапуске (особенность Counter метрик)
- **Визуализация**: Stacked bars (столбчатая диаграмма с группировкой)
- **Бизнес-значение**: Отслеживание успешности аутентификации, выявление проблем с логином

#### 3. Active Sessions (`sso_active_sessions`)
- **Тип**: Gauge
- **Описание**: Текущее количество активных пользовательских сессий (валидных JWT токенов)
- **Сохранность**: ℹ️ Обновляется динамически при каждом запросе с JWT
- **Визуализация**: Gauge (шкала)
- **Бизнес-значение**: Показывает реальную нагрузку на систему, помогает планировать ресурсы

### Дополнительные метрики

4. **Database Records** (`sso_database_records_total`)
   - Тип: Gauge
   - Описание: Количество записей в таблицах БД
   - Метки: `table` (users, audit)
   - Визуализация: Time series (smooth line)

5. **gRPC Calls** (`sso_grpc_calls_total`, `sso_grpc_call_duration_seconds`)
   - Тип: Counter, Histogram
   - Описание: Метрики вызовов к backend gRPC сервисам
   - Метки: `service` (admin, client, manager), `method`, `status`

## Запуск мониторинга

### С помощью Docker Compose

```bash
# Запуск всех сервисов (SSO, PostgreSQL, Prometheus, Grafana)
docker-compose up -d

# Просмотр логов
docker-compose logs -f sso
docker-compose logs -f prometheus
docker-compose logs -f grafana

# Остановка
docker-compose down

# Остановка с удалением данных
docker-compose down -v
```

### Локальный запуск (только для разработки)

```bash
# Убедитесь, что PostgreSQL запущен
# Запустите SSO API
go run ./cmd/sso

# В отдельных терминалах запустите Prometheus и Grafana
prometheus --config.file=deployments/prometheus/prometheus.yml
grafana-server --config=/etc/grafana/grafana.ini
```

## Доступ к интерфейсам

- **SSO API**: http://localhost:8081
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
  - Логин: `admin`
  - Пароль: `admin`

## Структура дашборда Grafana

Дашборд "SSO API Metrics Dashboard" включает:

### Верхний ряд (производительность)
1. **HTTP Request Rate** - График скорости запросов в секунду
   - Легенда: last, max, mean значения
   - Полезно для: Определения пиковых нагрузок

2. **HTTP Request Duration** - Перцентили времени ответа
   - Показывает p50 и p95
   - Легенда: last, max, mean значения
   - Полезно для: SLA мониторинга

### Средний ряд (текущее состояние)
3. **HTTP Requests In-Flight** - Gauge текущих запросов
   - Пороги: зеленый (0-5), желтый (5-10), красный (>10)

4. **Total Registered Users** - Gauge общего числа пользователей
   - Синий цвет, показывает текущее значение

5. **Active User Sessions** - Gauge активных сессий
   - Зеленый цвет, показывает текущее значение

### Нижний ряд (бизнес-метрики)
6. **Authentication Operations Rate** - Столбчатая диаграмма операций аутентификации
   - Зеленый: успешные операции
   - Красный: неудачные операции
   - Легенда: sum, last значения
   - Полезно для: Выявления проблем с логином

7. **Database Records Count** - График количества записей в БД
   - Smooth line визуализация
   - Легенда: last, max, min значения
   - Полезно для: Отслеживания роста данных

## Endpoints

### Metrics Endpoint

```bash
# Просмотр всех метрик в формате Prometheus
curl http://localhost:8081/metrics
```

Пример вывода:
```
# HELP sso_total_users Total number of registered users in the system
# TYPE sso_total_users gauge
sso_total_users 142

# HELP sso_auth_operations_total Total number of authentication operations
# TYPE sso_auth_operations_total counter
sso_auth_operations_total{operation="login",status="success"} 523
sso_auth_operations_total{operation="login",status="failure"} 12
sso_auth_operations_total{operation="signup",status="success"} 142
```

## Настройка алертов (опционально)

Рекомендуемые пороги для алертов:

```yaml
# prometheus/alerts.yml
groups:
  - name: sso_alerts
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: rate(sso_http_requests_total{status=~"5.."}[5m]) > 0.05
        annotations:
          summary: "High HTTP error rate detected"

      - alert: SlowRequests
        expr: histogram_quantile(0.95, rate(sso_http_request_duration_seconds_bucket[5m])) > 1
        annotations:
          summary: "95th percentile latency > 1s"

      - alert: HighAuthFailureRate
        expr: rate(sso_auth_operations_total{status="failure"}[5m]) > 0.1
        annotations:
          summary: "High authentication failure rate"
```

## Сохранность метрик при перезапуске

| Метрика | Тип | Сохранность | Пояснение |
|---------|-----|-------------|-----------|
| `sso_total_users` | Gauge | ✅ Да | Читается из БД при каждом обновлении |
| `sso_active_sessions` | Gauge | ⚠️ Частично | Обновляется при новых запросах с JWT |
| `sso_auth_operations_total` | Counter | ❌ Нет | Counter-метрики сбрасываются (используйте rate() в запросах) |
| `sso_database_records_total` | Gauge | ✅ Да | Читается из БД каждые 30 сек |
| `sso_http_requests_total` | Counter | ❌ Нет | Counter-метрики сбрасываются |

**Примечание**: Prometheus сохраняет исторические данные (retention: 7 дней), поэтому графики показывают правильные тренды даже после перезапуска приложения.

## Troubleshooting

### Метрики не отображаются в Grafana

1. Проверьте, что Prometheus scraping работает:
   ```bash
   curl http://localhost:9090/api/v1/targets
   ```

2. Проверьте доступность /metrics endpoint:
   ```bash
   curl http://localhost:8081/metrics
   ```

3. Проверьте логи Prometheus:
   ```bash
   docker-compose logs prometheus
   ```

### Grafana не подключается к Prometheus

1. Проверьте, что оба контейнера в одной сети:
   ```bash
   docker network inspect sso_sso-network
   ```

2. Проверьте datasource в Grafana UI (Configuration -> Data Sources)

## Полезные Prometheus запросы

```promql
# Request rate по методам
sum(rate(sso_http_requests_total[5m])) by (method)

# Топ медленных endpoints
topk(5, histogram_quantile(0.95, rate(sso_http_request_duration_seconds_bucket[5m])) by (path))

# Success rate аутентификации
sum(rate(sso_auth_operations_total{status="success"}[5m])) /
sum(rate(sso_auth_operations_total[5m]))

# Рост пользовательской базы за час
delta(sso_total_users[1h])
```

## Расширение метрик

Для добавления новой кастомной метрики:

1. Добавьте определение в [pkg/metrics/metrics.go](pkg/metrics/metrics.go:35)
2. Зарегистрируйте метрику в конструкторе `New()`
3. Вызывайте метрику в нужном месте кода
4. Обновите дашборд в [deployments/grafana/dashboards/sso-metrics.json](deployments/grafana/dashboards/sso-metrics.json:1)
5. Добавьте документацию в этот README

## Производительность

Метрики имеют минимальное влияние на производительность:
- Overhead на request: ~0.1-0.5ms
- Memory: ~10-20MB для метрик
- CPU: <1% при нормальной нагрузке

## Рекомендации

1. **Для продакшена**: Настройте remote storage для Prometheus (например, Thanos, Cortex)
2. **Безопасность**: Добавьте аутентификацию для /metrics endpoint
3. **Алерты**: Настройте Alertmanager для критичных метрик
4. **Retention**: Увеличьте retention Prometheus до 30-90 дней для продакшена
