package metrics

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type InfluxDBWriter struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	org      string
	bucket   string
}

type InfluxDBConfig struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

func NewInfluxDBWriter(cfg InfluxDBConfig) (*InfluxDBWriter, error) {
	client := influxdb2.NewClient(cfg.URL, cfg.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("influxdb connection failed: %w", err)
	}

	if health.Status != "pass" {
		return nil, fmt.Errorf("influxdb health check failed: status=%s", health.Status)
	}

	writeAPI := client.WriteAPI(cfg.Org, cfg.Bucket)

	return &InfluxDBWriter{
		client:   client,
		writeAPI: writeAPI,
		org:      cfg.Org,
		bucket:   cfg.Bucket,
	}, nil
}

func (w *InfluxDBWriter) WriteHTTPRequest(method, path, status string, duration float64, environment string) {
	p := influxdb2.NewPoint(
		"http_request",
		map[string]string{
			"method":      method,
			"path":        path,
			"status":      status,
			"service":     "sso-api",
			"environment": environment,
		},
		map[string]interface{}{
			"duration_seconds": duration,
			"count":            1,
		},
		time.Now(),
	)
	w.writeAPI.WritePoint(p)
}

func (w *InfluxDBWriter) WriteAuthOperation(operation, status, userRole, environment string, durationMs float64) {
	p := influxdb2.NewPoint(
		"auth_operation",
		map[string]string{
			"operation":   operation, // login, logout, refresh, register, password_reset
			"status":      status,    // success, failed, rate_limited
			"user_role":   userRole,  // admin, manager, client - MEANINGFUL TAG!
			"service":     "sso-api",
			"environment": environment,
		},
		map[string]interface{}{
			"count":        1,
			"duration_ms":  durationMs,
		},
		time.Now(),
	)
	w.writeAPI.WritePoint(p)
}


func (w *InfluxDBWriter) WriteTotalUsers(total int, activeCount int, newToday int, environment, region string) {
	p := influxdb2.NewPoint(
		"users",
		map[string]string{
			"service":     "sso-api",
			"environment": environment,
			"region":      region, 
		},
		map[string]interface{}{
			"total":           total,
			"active_last_24h": activeCount,
			"new_today":       newToday,
		},
		time.Now(),
	)
	w.writeAPI.WritePoint(p)
}

func (w *InfluxDBWriter) WritePasswordResetRequest(status, triggerSource, environment string, emailDuration float64) {
	p := influxdb2.NewPoint(
		"password_reset",
		map[string]string{
			"status":         status,         // requested, sent, failed, expired, completed
			"trigger_source": triggerSource,  // web, mobile, api
			"service":        "sso-api",
			"environment":    environment,
		},
		map[string]interface{}{
			"count":                   1,
			"email_send_duration_ms":  emailDuration,
		},
		time.Now(),
	)
	w.writeAPI.WritePoint(p)
}


func (w *InfluxDBWriter) WriteTokenMetrics(operation, status, environment string, validationDuration float64) {
	p := influxdb2.NewPoint(
		"jwt_token",
		map[string]string{
			"operation":   operation, // validate, refresh, revoke
			"status":      status,    // valid, expired, invalid, blacklisted
			"service":     "sso-api",
			"environment": environment,
		},
		map[string]interface{}{
			"count":                  1,
			"validation_duration_ms": validationDuration,
		},
		time.Now(),
	)
	w.writeAPI.WritePoint(p)
}


func (w *InfluxDBWriter) WriteActiveSessions(userRole, environment string, count int) {
	p := influxdb2.NewPoint(
		"active_sessions",
		map[string]string{
			"user_role":   userRole, // admin, manager, client
			"service":     "sso-api",
			"environment": environment,
		},
		map[string]interface{}{
			"count": count,
		},
		time.Now(),
	)
	w.writeAPI.WritePoint(p)
}


func (w *InfluxDBWriter) WriteDBQueryMetrics(queryType, table, environment string, duration float64, rowsAffected int64) {
	p := influxdb2.NewPoint(
		"db_query",
		map[string]string{
			"query_type":  queryType, // select, insert, update, delete
			"table":       table,
			"service":     "sso-api",
			"environment": environment,
		},
		map[string]interface{}{
			"duration_ms":    duration,
			"rows_affected":  rowsAffected,
		},
		time.Now(),
	)
	w.writeAPI.WritePoint(p)
}


func (w *InfluxDBWriter) WriteSuspiciousActivity(reason, severity, environment string, clientIPHash string) {
	p := influxdb2.NewPoint(
		"suspicious_activity",
		map[string]string{
			"reason":      reason,    // multiple_failed_logins, unknown_ip, token_reuse, sql_injection_attempt
			"severity":    severity,  // low, medium, high, critical
			"service":     "sso-api",
			"environment": environment,
		},
		map[string]interface{}{
			"count":          1,
			"client_ip_hash": clientIPHash,
		},
		time.Now(),
	)
	w.writeAPI.WritePoint(p)
}

func (w *InfluxDBWriter) WriteCustomMetric(measurement string, tags map[string]string, fields map[string]interface{}) {
	if tags == nil {
		tags = make(map[string]string)
	}
	tags["service"] = "sso-api"

	p := write.NewPoint(measurement, tags, fields, time.Now())
	w.writeAPI.WritePoint(p)
}

func (w *InfluxDBWriter) Flush() {
	w.writeAPI.Flush()
}

func (w *InfluxDBWriter) Close() {
	w.writeAPI.Flush()
	w.client.Close()
}

func (w *InfluxDBWriter) GetErrors() <-chan error {
	return w.writeAPI.Errors()
}
