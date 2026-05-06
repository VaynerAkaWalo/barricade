package oidc

import (
	"barricade/internal/identity"
	"barricade/internal/keys"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type (
	AccessToken string

	AccessTokenParams struct {
		Key           *keys.Key
		Ident         *identity.Identity
		ClientId      string
		Issuer        string
		Scope         string
		ExpiryMinutes int
	}
)

func NewAccessToken(params AccessTokenParams) (AccessToken, error) {
	privateKey, err := params.Key.RSAPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	now := time.Now()
	exp := now.Add(time.Duration(params.ExpiryMinutes) * time.Minute)

	claims := jwt.MapClaims{
		"iss":       params.Issuer,
		"sub":       string(params.Ident.Id),
		"aud":       params.Issuer,
		"exp":       exp.Unix(),
		"iat":       now.Unix(),
		"client_id": params.ClientId,
		"scope":     params.Scope,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = string(params.Key.Id)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return AccessToken(signedToken), nil
}
