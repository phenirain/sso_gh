package echomiddleware

import (
	// "context"

	// "github.com/getsentry/sentry-go"
	// sentryecho "github.com/getsentry/sentry-go/echo"
	// "github.com/labstack/echo/v4"
)

// PutSentryContext is an Echo middleware that extracts the Sentry hub from the Echo context
// and puts it into the standard context of the request. This allows other parts of the application
// to access the Sentry hub via the request context.
//TODO: разобраться с sentry
// func PutSentryContext(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		if hub := sentryecho.GetHubFromContext(c); hub != nil {
// 			hub.ConfigureScope(func(scope *sentry.Scope) {
// 				scope.SetRequest(c.Request())
// 			})
// 			ctx := c.Request().Context()
// 			ctx = context.WithValue(ctx, sentry.HubContextKey, hub)
// 			c.SetRequest(c.Request().WithContext(ctx))
// 		}
// 		return next(c)
// 	}
// }

