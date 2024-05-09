package admin

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type LoginClaims struct {
	Login   string `json:"login"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

var secret = os.Getenv("JWT_SECRET")

func EncodeLogin(login string, IsAdmin bool) (string, error) {
	claims := &LoginClaims{
		Login:   login,
		IsAdmin: IsAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func DecodeLogin(tokenString string) (*LoginClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &LoginClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		if claims, ok := token.Claims.(*LoginClaims); ok {
			return claims, nil
		}
	}
	return nil, fmt.Errorf("Token invalid")
}
