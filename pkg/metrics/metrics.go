package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	TotalUsersGauge prometheus.Gauge

	AuthOperationsTotal *prometheus.CounterVec

	InfluxDB *InfluxDBWriter
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

// RecordHTTPRequest records HTTP request metrics to both Prometheus and InfluxDB
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration float64) {
	// Prometheus metrics
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

	// InfluxDB metrics (if configured)
	if m.InfluxDB != nil {
		m.InfluxDB.WriteHTTPRequest(method, path, status, duration, "dev")
	}
}

// RecordAuthOperation records authentication operation
// userRole is a meaningful tag: admin, manager, or client
func (m *Metrics) RecordAuthOperation(operation, status, userRole string) {
	// Prometheus metrics
	m.AuthOperationsTotal.WithLabelValues(operation, status).Inc()

	// InfluxDB metrics with user_role tag (if configured)
	if m.InfluxDB != nil {
		m.InfluxDB.WriteAuthOperation(operation, status, userRole, "dev", 0)
	}
}

// SetTotalUsers sets the total users gauge
func (m *Metrics) SetTotalUsers(count float64) {
	m.TotalUsersGauge.Set(count)

	// InfluxDB metrics (if configured)
	if m.InfluxDB != nil {
		m.InfluxDB.WriteTotalUsers(int(count), 0, 0, "dev", "local")
	}
}

// Close closes all metric writers
func (m *Metrics) Close() {
	if m.InfluxDB != nil {
		m.InfluxDB.Close()
	}
}

