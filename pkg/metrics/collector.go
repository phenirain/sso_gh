package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

// Collector periodically collects metrics from database
type Collector struct {
	metrics *Metrics
	db      *sqlx.DB
	logger  *slog.Logger
}

// NewCollector creates new metrics collector
func NewCollector(m *Metrics, db *sqlx.DB, logger *slog.Logger) *Collector {
	return &Collector{
		metrics: m,
		db:      db,
		logger:  logger,
	}
}

// Start begins periodic metrics collection
func (c *Collector) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Collect immediately on start
	c.collectMetrics()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping metrics collector")
			return
		case <-ticker.C:
			c.collectMetrics()
		}
	}
}

func (c *Collector) collectMetrics() {
	// Collect total users count
	var userCount int64
	err := c.db.Get(&userCount, "SELECT COUNT(*) FROM users")
	if err != nil {
		c.logger.Error("Failed to collect user count metric", "error", err)
	} else {
		c.metrics.SetTotalUsers(float64(userCount))
	}

	// Collect audit records count
	var auditCount int64
	err = c.db.Get(&auditCount, "SELECT COUNT(*) FROM audit")
	if err != nil {
		c.logger.Error("Failed to collect audit count metric", "error", err)
	} else {
		c.metrics.SetDatabaseRecords("audit", float64(auditCount))
	}

	// Set users table count as well
	c.metrics.SetDatabaseRecords("users", float64(userCount))

	c.logger.Debug("Metrics collected",
		"users", userCount,
		"audit_records", auditCount,
	)
}
