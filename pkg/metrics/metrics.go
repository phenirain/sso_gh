package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all application metrics
type Metrics struct {
	// Standard HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Custom business metrics
	// 1. Total number of registered users in the system
	TotalUsersGauge prometheus.Gauge

	// 2. Authentication operations counter
	AuthOperationsTotal *prometheus.CounterVec

	// 3. Active sessions gauge (users currently authenticated)
	ActiveSessionsGauge prometheus.Gauge

	// Additional custom metrics
	// gRPC call metrics
	GRPCCallsTotal    *prometheus.CounterVec
	GRPCCallDuration  *prometheus.HistogramVec
	DatabaseRecordsGauge *prometheus.GaugeVec
}

// New creates and registers all metrics
func New() *Metrics {
	m := &Metrics{
		// Standard HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sso_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "sso_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "sso_http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		// Custom business metric 1: Total users in database
		TotalUsersGauge: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "sso_total_users",
				Help: "Total number of registered users in the system",
			},
		),

		// Custom business metric 2: Authentication operations
		AuthOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sso_auth_operations_total",
				Help: "Total number of authentication operations",
			},
			[]string{"operation", "status"}, // operation: login, signup, refresh, logout; status: success, failure
		),

		// Custom business metric 3: Active sessions
		ActiveSessionsGauge: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "sso_active_sessions",
				Help: "Current number of active user sessions (valid JWT tokens)",
			},
		),

		// gRPC call metrics
		GRPCCallsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sso_grpc_calls_total",
				Help: "Total number of gRPC calls to backend services",
			},
			[]string{"service", "method", "status"}, // service: admin, client, manager
		),
		GRPCCallDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "sso_grpc_call_duration_seconds",
				Help:    "gRPC call duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		),

		// Database records gauge
		DatabaseRecordsGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sso_database_records_total",
				Help: "Total number of records in database tables",
			},
			[]string{"table"}, // table: users, audit, etc.
		),
	}

	return m
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordAuthOperation records authentication operation
func (m *Metrics) RecordAuthOperation(operation, status string) {
	m.AuthOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordGRPCCall records gRPC call metrics
func (m *Metrics) RecordGRPCCall(service, method, status string, duration float64) {
	m.GRPCCallsTotal.WithLabelValues(service, method, status).Inc()
	m.GRPCCallDuration.WithLabelValues(service, method).Observe(duration)
}

// SetTotalUsers sets the total users gauge
func (m *Metrics) SetTotalUsers(count float64) {
	m.TotalUsersGauge.Set(count)
}

// SetActiveSessions sets the active sessions gauge
func (m *Metrics) SetActiveSessions(count float64) {
	m.ActiveSessionsGauge.Set(count)
}

// SetDatabaseRecords sets the database records gauge for a table
func (m *Metrics) SetDatabaseRecords(table string, count float64) {
	m.DatabaseRecordsGauge.WithLabelValues(table).Set(count)
}

// IncrementInFlight increments in-flight requests
func (m *Metrics) IncrementInFlight() {
	m.HTTPRequestsInFlight.Inc()
}

// DecrementInFlight decrements in-flight requests
func (m *Metrics) DecrementInFlight() {
	m.HTTPRequestsInFlight.Dec()
}
