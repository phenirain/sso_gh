package auth

// AuthRequest содержит учетные данные для авторизации
// swagger:model AuthRequest
type AuthRequest struct {
	// Логин пользователя
	Login string `json:"login" example:"user@example.com"`
	// Пароль пользователя
	Password string `json:"password" example:"P@ssw0rd!"`
}

// AuthResponse возвращает JWT токены после успешной аутентификации
// swagger:model AuthResponse
type AuthResponse struct {
	// Refresh Token для обновления пары токенов
	RefreshToken string `json:"refresh_token"`
	// Access Token для доступа к защищенным ресурсам
	AccessToken string `json:"access_token"`
	
}
