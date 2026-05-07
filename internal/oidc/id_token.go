package oidc

import (
	"barricade/internal/identity"
	"barricade/internal/keys"
	"crypto/sha256"
	"encoding/base64"
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
		Nonce         string
		ExpiryMinutes int
		AuthTime      int64
		AccessToken   string
	}
)

func atHash(accessToken string) string {
	hash := sha256.Sum256([]byte(accessToken))
	return base64.RawURLEncoding.EncodeToString(hash[:16])
}

func NewIdToken(params IdTokenParams) (IdToken, error) {
	privateKey, err := params.Key.RSAPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	now := time.Now()
	exp := now.Add(time.Duration(params.ExpiryMinutes) * time.Minute)

	claims := jwt.MapClaims{
		"iss":       params.Issuer,
		"sub":       string(params.Ident.Id),
		"aud":       params.ClientId,
		"exp":       exp.Unix(),
		"iat":       now.Unix(),
		"auth_time": params.AuthTime,
		"acr":       "0",
		"amr":       []string{"pwd"},
	}

	if params.AccessToken != "" {
		claims["at_hash"] = atHash(params.AccessToken)
	}

	if params.Nonce != "" {
		claims["nonce"] = params.Nonce
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = string(params.Key.Id)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return IdToken(signedToken), nil
}
