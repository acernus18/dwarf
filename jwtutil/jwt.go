package jwtutil

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
)

func GetToken(secret string, claims jwt.RegisteredClaims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseToken(secret, token string) (*jwt.RegisteredClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if tokenClaims == nil {
		return nil, errors.New("[110000]: parse token fail")
	}
	if result, ok := tokenClaims.Claims.(*jwt.RegisteredClaims); ok && tokenClaims.Valid {
		return result, nil
	}
	return nil, errors.New("[110000]: parse token fail")
}
