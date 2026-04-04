package oidc

import (
	"barricade/internal/identity"
	"barricade/internal/keys"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type (
	IdToken string

	IdTokenParams struct {
		Key           *keys.Key
		Ident         *identity.Identity
		ClientId      string
		Issuer        string
		ExpiryMinutes int
	}
)

func NewIdToken(params IdTokenParams) (IdToken, error) {
	privateKey, err := params.Key.RSAPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	now := time.Now()
	exp := now.Add(time.Duration(params.ExpiryMinutes) * time.Minute)

	claims := jwt.MapClaims{
		"iss": params.Issuer,
		"sub": string(params.Ident.Id),
		"aud": params.ClientId,
		"exp": exp.Unix(),
		"iat": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = string(params.Key.Id)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return IdToken(signedToken), nil
}
