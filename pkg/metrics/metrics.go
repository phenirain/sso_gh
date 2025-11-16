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
}

func New() *Metrics {
	m := &Metrics{
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

		TotalUsersGauge: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "sso_total_users",
				Help: "Total number of registered users in the system",
			},
		),

		AuthOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sso_auth_operations_total",
				Help: "Total number of authentication operations",
			},
			[]string{"operation", "status"},
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

// SetTotalUsers sets the total users gauge
func (m *Metrics) SetTotalUsers(count float64) {
	m.TotalUsersGauge.Set(count)
}

