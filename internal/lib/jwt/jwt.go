package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	jwtErrors "github.com/phenirain/sso/internal/errors/jwt"
	"time"
)

type JwtLib struct {
	duration time.Duration
	secret []byte
}

func NewJwtLib(duration time.Duration, secret []byte) *JwtLib {
	return &JwtLib{
		duration: duration,
		secret: secret,
	}
}

func (j *JwtLib) NewToken(userId, role int64) (accessToken string, refreshToken string, error error) {
	claims := jwt.MapClaims{
		"sub":  userId,
		"role": role,
	}
	claims["exp"] = time.Now().Add(j.duration).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(j.secret)
	if err != nil {
		return "", "", err
	}

	claims["exp"] = time.Now().Add(time.Hour*24*30).Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err = token.SignedString(j.secret)
	if err != nil {
		return "", "", err
	}
	return
}

func (j *JwtLib) ParseToken(tokenString string) (userId int64, roleId int64, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return -1, -1, fmt.Errorf("token parse error: %s", err.Error())
	}
	if !token.Valid {
		return -1, -1, jwtErrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return -1, -1, errors.New("can't get claims")
	}

	uid, ok := claims["sub"]
	if !ok {
		return -1, -1, errors.New("can't get sub from claims")
	}

	role, ok := claims["role"]
	if !ok {
		return -1, -1, errors.New("can't get role from claims")
	}

	return int64(uid.(float64)), int64(role.(float64)), nil
}
