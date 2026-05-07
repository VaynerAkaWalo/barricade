package oidc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"barricade/internal/identity"
	"barricade/internal/keys"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdTokenTest(t *testing.T) (*keys.Key, *identity.Identity) {
	repo := keys.NewInMemoryRepository()
	service := keys.NewService(repo)

	key, err := service.CreateKey(context.Background(), keys.RS256)
	require.NoError(t, err)

	ident, err := identity.New("test-user", "secret123")
	require.NoError(t, err)

	return key, ident
}

func TestNewIdTokenWithAllClaims(t *testing.T) {
	key, ident := setupIdTokenTest(t)
	authTime := time.Now().Unix()

	accessToken := "test-access-token-value"
	expectedHash := sha256.Sum256([]byte(accessToken))
	expectedAtHash := base64.RawURLEncoding.EncodeToString(expectedHash[:16])

	idToken, err := NewIdToken(IdTokenParams{
		Key:           key,
		Ident:         ident,
		ClientId:      "test-client-id",
		Issuer:        "https://issuer.example.com",
		Nonce:         "test-nonce",
		ExpiryMinutes: 5,
		AuthTime:      authTime,
		AccessToken:   accessToken,
	})
	require.NoError(t, err)
	require.NotEmpty(t, idToken)

	token, _, err := jwt.NewParser().ParseUnverified(string(idToken), jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, "https://issuer.example.com", claims["iss"])
	assert.Equal(t, string(ident.Id), claims["sub"])
	assert.Equal(t, "test-client-id", claims["aud"])
	assert.Equal(t, "test-nonce", claims["nonce"])

	authTimeClaim, ok := claims["auth_time"].(float64)
	assert.True(t, ok)
	assert.Equal(t, authTime, int64(authTimeClaim))

	assert.Equal(t, "0", claims["acr"])
	assert.Equal(t, []interface{}{"pwd"}, claims["amr"])

	assert.Equal(t, expectedAtHash, claims["at_hash"])

	expClaim, ok := claims["exp"].(float64)
	assert.True(t, ok)
	assert.Greater(t, expClaim, float64(time.Now().Unix()))

	iatClaim, ok := claims["iat"].(float64)
	assert.True(t, ok)
	assert.LessOrEqual(t, iatClaim, float64(time.Now().Unix()))
}

func TestNewIdTokenWithoutNonceAndAccessToken(t *testing.T) {
	key, ident := setupIdTokenTest(t)

	idToken, err := NewIdToken(IdTokenParams{
		Key:           key,
		Ident:         ident,
		ClientId:      "test-client-id",
		Issuer:        "https://issuer.example.com",
		Nonce:         "",
		ExpiryMinutes: 5,
		AuthTime:      0,
		AccessToken:   "",
	})
	require.NoError(t, err)

	token, _, err := jwt.NewParser().ParseUnverified(string(idToken), jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	_, hasNonce := claims["nonce"]
	assert.False(t, hasNonce)

	_, hasAtHash := claims["at_hash"]
	assert.False(t, hasAtHash)

	assert.Equal(t, "0", claims["acr"])
	assert.Equal(t, []interface{}{"pwd"}, claims["amr"])

	authTimeClaim, ok := claims["auth_time"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(0), authTimeClaim)
}

func TestAtHash(t *testing.T) {
	accessToken := "my-access-token"
	hash := sha256.Sum256([]byte(accessToken))
	expected := base64.RawURLEncoding.EncodeToString(hash[:16])

	result := atHash(accessToken)
	assert.Equal(t, expected, result)
}

func TestAtHashDeterministic(t *testing.T) {
	token := "same-token-value"
	assert.Equal(t, atHash(token), atHash(token))
}

func TestNewIdTokenWithKidHeader(t *testing.T) {
	key, ident := setupIdTokenTest(t)

	idToken, err := NewIdToken(IdTokenParams{
		Key:           key,
		Ident:         ident,
		ClientId:      "client-id",
		Issuer:        "https://issuer.example.com",
		ExpiryMinutes: 5,
		AuthTime:      0,
	})
	require.NoError(t, err)

	token, _, err := jwt.NewParser().ParseUnverified(string(idToken), jwt.MapClaims{})
	require.NoError(t, err)

	kid, ok := token.Header["kid"].(string)
	assert.True(t, ok)
	assert.Equal(t, string(key.Id), kid)
}
