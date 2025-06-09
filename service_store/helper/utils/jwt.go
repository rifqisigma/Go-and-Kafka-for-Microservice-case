package utils

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var jwt_secret = []byte(os.Getenv("JWT_SECRET"))

type JWTCLAIMS struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func ValidateJWT(tokenstring string) (*JWTCLAIMS, error) {
	token, err := jwt.ParseWithClaims(tokenstring, &JWTCLAIMS{}, func(t *jwt.Token) (interface{}, error) {
		return jwt_secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTCLAIMS)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
