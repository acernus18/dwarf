package jwtutil

import (
	"github.com/golang-jwt/jwt/v4"
	"testing"
)

func TestToken(t *testing.T) {
	claims := jwt.RegisteredClaims{
		Issuer:    "Maples",
		Subject:   "",
		Audience:  []string{"111", "22"},
		ExpiresAt: nil,
		NotBefore: nil,
		IssuedAt:  nil,
		ID:        "",
	}

	token, err := GetToken("Maple", claims)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
	result, err := ParseToken("Maple", token)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}
