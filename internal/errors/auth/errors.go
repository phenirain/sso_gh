package auth

import "errors"

var (
	ErrInvalidUserCredentials = errors.New("неверен логин или пароль")
	ErrUserAlreadyExists      = errors.New("пользователь уже существует")
	ErrUserNotFound           = errors.New("пользователь не существует")
	ErrUserArchived           = errors.New("Ваш аккаунт удален, напишите письмо на почту \"phenirain@gmail.com\"")
)
