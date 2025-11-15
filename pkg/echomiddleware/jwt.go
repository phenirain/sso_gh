package echomiddleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/pkg/contextkeys"
)

type Jwt interface {
	ParseToken(tokenString string) (userId int64, roleId int64, err error)
}

const (
	RoleClient  int64 = 1
	RoleManager int64 = 2
	RoleAdmin   int64 = 3
)

func JwtValidation(jwt Jwt) echo.MiddlewareFunc {
	skip := map[string]struct{}{
		"/auth/logIn":   {},
		"/auth/signUp":  {},
		"/auth/refresh": {},
		"/health":       {},
		"/swagger/*":    {},
		"/v":            {},
		"/metrics":      {},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Пропускаем OPTIONS запросы (CORS preflight)
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}

			p := c.Path()
			if _, ok := skip[p]; ok {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.ErrUnauthorized
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.ErrUnauthorized
			}
			tokenString := parts[1]

			userId, roleId, err := jwt.ParseToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": err.Error(),
				})
			}

			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, contextkeys.UserIDCtxKey, userId)
			ctx = context.WithValue(ctx, contextkeys.RoleIDCtxKey, roleId)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// RoleMiddleware проверяет, что роль пользователя соответствует разрешенным ролям
func RoleMiddleware(allowedRoles ...int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleId, ok := c.Request().Context().Value(contextkeys.RoleIDCtxKey).(int64)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "role not found in context",
				})
			}

			// Проверяем, есть ли роль пользователя в списке разрешенных
			for _, allowedRole := range allowedRoles {
				if roleId == allowedRole {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "access denied: insufficient permissions",
			})
		}
	}
}
