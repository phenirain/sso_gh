package echomiddleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phenirain/sso/pkg/contextkeys"
)

const (
	parentSeparatorNumber = 3 // https://www.w3.org/TR/trace-context/#version-format
	// traceparent: {version}-{trace-id}-{parent-id}-{trace-flags}
)

type logger interface {
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
}

// SlogLoggerMiddleware returns an Echo middleware that logs requests using slog.
func SlogLoggerMiddleware(logger logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogRemoteIP: true,
		LogProtocol: true,
		LogUserAgent: true,
		LogLatency: true,
		LogError:   true,
		LogHeaders: []string{RequestIDHeader, TraceParentHeader},
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []any{
				slog.String("path", v.URI),
				slog.Int("status_code", v.Status),
				slog.String("method", v.Method),
				slog.String("method", v.Method),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("user_agent", v.UserAgent),
				slog.String("exec_time", v.Latency.String()),
			}
			
			level := slog.LevelInfo
			msg := "REQUEST"
			// reqID := getRequestID(v.Headers)
			// attrs = append(attrs, slog.String(string(contextkeys.RequestIDCtxKey), reqID))
			// traceID := getTraceID(v.Headers)
			// attrs = append(attrs, slog.String(string(contextkeys.TraceIDCtxKey), traceID))
			userID := c.Get(string(contextkeys.UserIDCtxKey))
			attrs = append(attrs, slog.String(string(contextkeys.UserIDCtxKey), userID.(string)))

			respErrStr := "?"
			if v.Error != nil {
				attrs = append(attrs, slog.String("err", v.Error.Error()))
			}

			// Change level on 5xx
			if v.Status >= http.StatusInternalServerError {
				level = slog.LevelError
				msg = "REQUEST_ERROR: " + respErrStr
			}
			logger.LogAttrs(c.Request().Context(), level, msg, slog.Group("context", attrs...))
			return nil
		},
	})
}