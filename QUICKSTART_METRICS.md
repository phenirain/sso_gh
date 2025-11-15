# Быстрый старт - Мониторинг SSO API

## 🚀 Запуск мониторинга

### 1. Запустите все сервисы

```bash
cd deployments
docker-compose up -d
```

Это запустит:
- ✅ SSO API на порту **8081**
- ✅ Prometheus на порту **9090**
- ✅ Grafana на порту **3000**

### 2. Проверьте статус

```bash
# Проверка запущенных контейнеров
docker-compose ps

# Просмотр логов
docker-compose logs -f sso
```

### 3. Откройте интерфейсы

| Сервис | URL | Логин/Пароль |
|--------|-----|--------------|
| **SSO API** | http://localhost:8081 | - |
| **Swagger** | http://localhost:8081/swagger/index.html | - |
| **Метрики** | http://localhost:8081/metrics | - |
| **Prometheus** | http://localhost:9090 | - |
| **Grafana** | http://localhost:3000 | admin/admin |

### 4. Просмотр метрик в Grafana

1. Откройте http://localhost:3000
2. Войдите с логином `admin` и паролем `admin`
3. Перейдите в **Dashboards** → **SSO API Metrics Dashboard**
4. Вы увидите все метрики в реальном времени!

## 📊 Метрики приложения

### Кастомные бизнес-метрики:

1. **Total Users** (`sso_total_users`)
   - Общее количество зарегистрированных пользователей
   - ✅ Данные сохраняются при перезапуске (читаются из БД)

2. **Authentication Operations** (`sso_auth_operations_total`)
   - Счетчик операций аутентификации (login, signup, refresh)
   - Метки: operation, status (success/failure)

3. **Active Sessions** (`sso_active_sessions`)
   - Текущее количество активных пользовательских сессий

### Стандартные метрики:

- HTTP Request Rate
- HTTP Request Duration (p50, p95)
- HTTP Requests In-Flight
- Database Records Count
- gRPC Calls

## 🧪 Тестирование метрик

### Генерация тестовой нагрузки

```bash
# Регистрация нового пользователя
curl -X POST http://localhost:8081/auth/signUp \
  -H "Content-Type: application/json" \
  -d '{"login": "testuser", "password": "testpass"}'

# Логин
curl -X POST http://localhost:8081/auth/logIn \
  -H "Content-Type: application/json" \
  -d '{"login": "testuser", "password": "testpass"}'

# Просмотр всех метрик
curl http://localhost:8081/metrics
```

### Проверка конкретных метрик

```bash
# Количество пользователей
curl -s http://localhost:8081/metrics | grep sso_total_users

# Операции аутентификации
curl -s http://localhost:8081/metrics | grep sso_auth_operations_total

# Активные сессии
curl -s http://localhost:8081/metrics | grep sso_active_sessions
```

## 📈 Визуализации в Grafana

Дашборд содержит 7 панелей:

1. **HTTP Request Rate** - График скорости запросов (Time Series)
2. **HTTP Request Duration** - Перцентили времени ответа (Time Series)
3. **HTTP Requests In-Flight** - Текущие запросы (Gauge)
4. **Total Registered Users** - Количество пользователей (Gauge)
5. **Active User Sessions** - Активные сессии (Gauge)
6. **Authentication Operations Rate** - Операции аутентификации (Stacked Bars)
7. **Database Records Count** - Записи в БД (Time Series)

## 🛑 Остановка сервисов

```bash
# Остановить все сервисы
docker-compose down

# Остановить и удалить данные (volumes)
docker-compose down -v
```

## 🔍 Troubleshooting

### Метрики не отображаются?

```bash
# 1. Проверьте /metrics endpoint
curl http://localhost:8081/metrics

# 2. Проверьте targets в Prometheus
curl http://localhost:9090/api/v1/targets | jq

# 3. Проверьте логи
docker-compose logs prometheus
docker-compose logs grafana
```

### Grafana не показывает данные?

1. Убедитесь, что datasource Prometheus подключен:
   - Configuration → Data Sources → Prometheus
   - Должен быть URL: `http://prometheus:9090`

2. Проверьте, что SSO контейнер доступен:
   ```bash
   docker exec -it sso-prometheus wget -O- http://sso:8081/metrics
   ```

## 📚 Дополнительная информация

Подробная документация: [MONITORING.md](../MONITORING.md)

## ✅ Чеклист выполнения задания

- [x] Выгрузка данных в формате Prometheus (`/metrics` endpoint)
- [x] 3 кастомные метрики (users, auth operations, active sessions)
- [x] Визуализация метрик в Grafana (7 панелей)
- [x] Логичные типы визуализации (Gauge, Time Series, Bars)
- [x] Легенды с основными данными (last, max, mean, sum, min)
- [x] Метрики сохраняются при перезапуске (через БД для Gauge метрик)
