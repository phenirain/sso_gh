package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/phenirain/sso/internal/config"
	"github.com/phenirain/sso/internal/domain"
	"github.com/phenirain/sso/internal/dto/auth"
	authErrors "github.com/phenirain/sso/internal/errors/auth"
	"github.com/phenirain/sso/internal/errors/jwt"
	"github.com/phenirain/sso/pkg/contextkeys"
	api "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client"
)

type Jwt interface {
	NewToken(userId, role int64) (accessToken string, refreshToken string, error error)
	ParseToken(tokenString string) (userId int64, roleId int64, err error)
}

type Repository interface {
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
	GetUserWithId(ctx context.Context, uid int64) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	UpdatePassword(ctx context.Context, login, newPasswordHash string) error
}

type Auth struct {
	s      pb.ClientServiceClient
	repo   Repository
	jwt    Jwt
	config *config.Config
}

func New(repo Repository, jwt Jwt, clientService pb.ClientServiceClient, cfg *config.Config) *Auth {
	return &Auth{
		repo:   repo,
		jwt:    jwt,
		s:      clientService,
		config: cfg,
	}
}

func (a *Auth) Auth(ctx context.Context, request auth.AuthRequest, isNew bool) (*auth.AuthResponse, error) {
	const op string = "Auth.Login"

	user, err := a.repo.GetUserByLogin(ctx, request.Login)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var userId int64
	var role int64
	// если создание
	if isNew {
		// если пользователь найден - уже существует
		if user != nil {
			return nil, authErrors.ErrUserAlreadyExists
		}

		user = domain.NewUser(request.Login, request.Password, nil, nil)
		role = 1
		userId, err = a.repo.CreateUser(ctx, user)
		if err != nil {
			errText := fmt.Errorf("ошибка в ходе создания пользователя: %w", err)
			slog.Error(errText.Error())
			return nil, errText
		}

		req := api.ClientRequest{
			Email:  &request.Login,
			UserId: &userId,
		}

		ctx = context.WithValue(ctx, contextkeys.UserIDCtxKey, userId)
		_, err = a.s.RegisterClient(ctx, &req)
		if err != nil {
			return nil, fmt.Errorf("ошибка регистрации клиента: %w", err)
		}
	} else { // если авторизация
		// если пользователь не найден
		if user == nil {
			return nil, authErrors.ErrInvalidUserCredentials
		}
		// проверяем, не архивирован ли пользователь
		if user.IsArchived {
			return nil, authErrors.ErrUserArchived
		}
		valid := user.CheckPassword(request.Password)
		// если пароль не верен
		if !valid {
			return nil, authErrors.ErrInvalidUserCredentials
		}
		userId = user.Id
		role = user.RoleId
	}

	return a.getAuthResponse(userId, role)
}

func (a *Auth) Refresh(ctx context.Context, refreshToken string) (*auth.AuthResponse, error) {

	// проверка токена
	userId, roleId, err := a.jwt.ParseToken(refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrInvalidToken) {
			return nil, err
		}
		slog.Error("ошибка парсинга токена", "err", err)
		return nil, err
	}

	// проверка пользователя
	user, err := a.repo.GetUserWithId(ctx, userId)
	if err != nil {
		errorText := fmt.Errorf("ошибка получения пользователя по идентфикатору: %w", err)
		slog.Error(errorText.Error())
		return nil, errorText
	}
	// если его нет или удален - нахуй
	if user == nil || user.IsArchived {
		return nil, authErrors.ErrUserNotFound
	}

	// Используем роль из токена, но можно проверить соответствие с ролью в БД
	if roleId != user.RoleId {
		slog.Warn("role mismatch between token and database", "tokenRole", roleId, "dbRole", user.RoleId)
		// Используем роль из БД как источник истины
		roleId = user.RoleId
	}

	return a.getAuthResponse(userId, roleId)
}

func (a *Auth) getAuthResponse(userId, role int64) (*auth.AuthResponse, error) {
	accessToken, refreshToken, err := a.jwt.NewToken(userId, role)
	if err != nil {
		errorText := fmt.Errorf("ошибка генерации токенов доступа: %w", err)
		slog.Error(errorText.Error())
		return nil, errorText
	}

	return &auth.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		RoleId:       role,
	}, nil
}

func (a *Auth) SendPasswordResetEmail(ctx context.Context, login string) error {
	const op = "Auth.SendPasswordResetEmail"

	// Проверяем существование пользователя
	user, err := a.repo.GetUserByLogin(ctx, login)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	if user == nil {
		return authErrors.ErrUserNotFound
	}

	// Проверяем, не архивирован ли пользователь
	if user.IsArchived {
		return authErrors.ErrUserArchived
	}

	// Email это login пользователя (используется при регистрации)
	userEmail := login

	// Кодируем логин в base64
	encodedLogin := base64.StdEncoding.EncodeToString([]byte(login))

	// Формируем ссылку для сброса пароля
	resetLink := fmt.Sprintf("%s?token=%s", a.config.Email.FrontendResetURL, encodedLogin)

	// Отправляем запрос на email-сервис
	emailPayload := map[string]interface{}{
		"to":        userEmail,
		"resetLink": resetLink,
		"login":     login,
	}

	jsonData, err := json.Marshal(emailPayload)
	if err != nil {
		return fmt.Errorf("%s: ошибка создания JSON: %w", op, err)
	}

	resp, err := http.Post(
		a.config.Email.ServiceURL+"/send-reset-email",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		slog.Error("failed to send email", "err", err)
		return fmt.Errorf("%s: ошибка отправки email: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: email-сервис вернул статус %d", op, resp.StatusCode)
	}

	slog.Info("password reset email sent", "login", login, "email", userEmail)
	return nil
}

func (a *Auth) ResetPassword(ctx context.Context, login, newPassword string) error {
	const op = "Auth.ResetPassword"

	// Проверяем существование пользователя
	user, err := a.repo.GetUserByLogin(ctx, login)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	if user == nil {
		return authErrors.ErrUserNotFound
	}

	// Проверяем, не архивирован ли пользователь
	if user.IsArchived {
		return authErrors.ErrUserArchived
	}

	// Хешируем новый пароль используя ту же логику что и при создании
	newUser := domain.NewUser(login, newPassword, nil, nil)

	// Обновляем пароль (PasswordHash это []byte, нужно string)
	err = a.repo.UpdatePassword(ctx, login, string(newUser.PasswordHash))
	if err != nil {
		slog.Error("failed to update password", "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	slog.Info("password reset successful", "login", login)
	return nil
}
