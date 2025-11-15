package echomiddleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/pkg/metrics"
)

// MetricsMiddleware creates middleware for tracking HTTP metrics
func MetricsMiddleware(m *metrics.Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip metrics endpoint to avoid recursion
			if c.Path() == "/metrics" {
				return next(c)
			}

			// Increment in-flight requests
			m.IncrementInFlight()
			defer m.DecrementInFlight()

			start := time.Now()

			// Process request
			err := next(c)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := c.Response().Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = 500
				}
			}

			method := c.Request().Method
			path := c.Path()
			statusStr := strconv.Itoa(status)

			m.RecordHTTPRequest(method, path, statusStr, duration)

			return err
		}
	}
}
