package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	authModels "github.com/phenirain/sso/internal/dto/auth"
	"github.com/phenirain/sso/internal/dto/response"
	"github.com/phenirain/sso/pkg/metrics"
)

type AuthService interface {
	Auth(ctx context.Context, request authModels.AuthRequest, isNew bool) (*authModels.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*authModels.AuthResponse, error)
	ResetPassword(ctx context.Context, login, newPassword string) error
	SendPasswordResetEmail(ctx context.Context, login string) error
}

type Handler struct {
	s AuthService
	m *metrics.Metrics
}

func NewHandler(auth AuthService, m *metrics.Metrics) *Handler {
	return &Handler{
		s: auth,
		m: m,
	}
}

// LogIn godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authModels.AuthRequest true "Credentials"
// @Success 200 {object} authModels.AuthResponse
// @Router /auth/logIn [post]
func (h *Handler) LogIn(c echo.Context) error {
	return h.auth(c, false)
}

// SignUp godoc
// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authModels.AuthRequest true "Credentials"
// @Success 200 {object} authModels.AuthResponse
// @Router /auth/signUp [post]
func (h *Handler) SignUp(c echo.Context) error {
	return h.auth(c, true)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags auth
// @Produce json
// @Success 200 {object} authModels.AuthResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c echo.Context) error {
	ctx := c.Request().Context()

	// Получаем токен из заголовка Authorization
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		h.m.RecordAuthOperation("refresh", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Отсутствует токен", "Заголовок Authorization обязателен"))
	}

	// Проверяем формат "Bearer <token>"
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		h.m.RecordAuthOperation("refresh", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Неверный формат токена", "Используйте формат: Bearer <token>"))
	}

	refreshToken := authHeader[7:] // Убираем "Bearer "

	result, err := h.s.Refresh(ctx, refreshToken)
	if err != nil {
		h.m.RecordAuthOperation("refresh", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка обновления токена", err.Error()))
	}

	h.m.RecordAuthOperation("refresh", "success")
	return c.JSON(http.StatusOK, response.NewSuccessResponse(result))
}

// ForgotPasswordRequest represents the request body for forgot password
type ForgotPasswordRequest struct {
	Login string `json:"login" example:"user@example.com"`
}

// ForgotPassword godoc
// @Summary Send password reset email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "User login"
// @Success 200 {object} response.ApiResponse[any]
// @Router /auth/forgotPassword [post]
func (h *Handler) ForgotPassword(c echo.Context) error {
	ctx := c.Request().Context()

	var req ForgotPasswordRequest

	if err := c.Bind(&req); err != nil {
		h.m.RecordAuthOperation("forgot_password", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	if req.Login == "" {
		h.m.RecordAuthOperation("forgot_password", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Отсутствует аргумент", "Логин обязателен"))
	}

	err := h.s.SendPasswordResetEmail(ctx, req.Login)
	if err != nil {
		h.m.RecordAuthOperation("forgot_password", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка отправки письма", err.Error()))
	}

	h.m.RecordAuthOperation("forgot_password", "success")
	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Письмо для сброса пароля отправлено на почту"))
}

// ResetPassword godoc
// @Summary Reset user password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authModels.AuthRequest true "Login and new password"
// @Success 200 {object} response.ApiResponse[any]
// @Router /auth/resetPassword [post]
func (h *Handler) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()

	var req authModels.AuthRequest
	if err := c.Bind(&req); err != nil {
		h.m.RecordAuthOperation("password_reset", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	if req.Login == "" {
		h.m.RecordAuthOperation("password_reset", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Отсутствует аргумент", "Логин обязателен"))
	}
	if req.Password == "" {
		h.m.RecordAuthOperation("password_reset", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Отсутствует аргумент", "Новый пароль обязателен"))
	}

	// Декодируем login из base64 если он закодирован
	decodedLogin := req.Login
	if decoded, err := base64.StdEncoding.DecodeString(req.Login); err == nil {
		decodedLogin = string(decoded)
	}

	err := h.s.ResetPassword(ctx, decodedLogin, req.Password)
	if err != nil {
		h.m.RecordAuthOperation("password_reset", "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка сброса пароля", err.Error()))
	}

	h.m.RecordAuthOperation("password_reset", "success")
	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty(fmt.Sprintf("Пароль для пользователя %s успешно изменен", decodedLogin)))
}

func (h *Handler) auth(c echo.Context, isNew bool) error {
	ctx := c.Request().Context()

	operation := "login"
	if isNew {
		operation = "signup"
	}

	var req authModels.AuthRequest
	if err := c.Bind(&req); err != nil {
		h.m.RecordAuthOperation(operation, "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	if req.Login == "" {
		h.m.RecordAuthOperation(operation, "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Отсутствует аргумент", "Логин обязателен"))
	}
	if req.Password == "" {
		h.m.RecordAuthOperation(operation, "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Отсутствует аргумент", "Пароль обязателен"))
	}

	result, err := h.s.Auth(ctx, req, isNew)
	if err != nil {
		h.m.RecordAuthOperation(operation, "failure")
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка авторизации", err.Error()))
	}

	h.m.RecordAuthOperation(operation, "success")
	return c.JSON(http.StatusOK, response.NewSuccessResponse(result))
}
