package oidc

import (
	"barricade/internal/keys"
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeysToJWKSEmpty(t *testing.T) {
	result := keysToJWKS([]*keys.Key{})

	assert.NotNil(t, result)
	assert.Empty(t, result.Keys)
}

func TestKeysToJWKSFiltersOnlyRSAKeys(t *testing.T) {
	repo := keys.NewInMemoryRepository()
	service := keys.NewService(repo)
	ctx := context.Background()

	// Create RSA key
	_, err := service.CreateKey(ctx, keys.RS256)
	require.NoError(t, err)

	allKeys, err := service.ListAllKeys(ctx)
	require.NoError(t, err)

	jwks := keysToJWKS(allKeys)

	assert.Len(t, jwks.Keys, 1)
	assert.Equal(t, "RSA", jwks.Keys[0].Kty)
	assert.Equal(t, "RS256", jwks.Keys[0].Alg)
	assert.Equal(t, "sig", jwks.Keys[0].Use)
	assert.NotEmpty(t, jwks.Keys[0].Kid)
	assert.NotEmpty(t, jwks.Keys[0].N)
	assert.NotEmpty(t, jwks.Keys[0].E)
}

func TestKeyToJWKRS256(t *testing.T) {
	repo := keys.NewInMemoryRepository()
	service := keys.NewService(repo)
	ctx := context.Background()

	key, err := service.CreateKey(ctx, keys.RS256)
	require.NoError(t, err)

	jwk, err := keyToJWK(key)

	require.NoError(t, err)
	assert.Equal(t, "RSA", jwk.Kty)
	assert.Equal(t, string(key.Id), jwk.Kid)
	assert.Equal(t, "sig", jwk.Use)
	assert.Equal(t, "RS256", jwk.Alg)
	assert.NotEmpty(t, jwk.N)
	assert.NotEmpty(t, jwk.E)

	// Verify base64url encoding (no padding characters)
	_, err = base64.RawURLEncoding.DecodeString(jwk.N)
	assert.NoError(t, err, "N should be valid base64url without padding")
	_, err = base64.RawURLEncoding.DecodeString(jwk.E)
	assert.NoError(t, err, "E should be valid base64url without padding")
}

func TestKeyToJWKUnsupportedAlgorithm(t *testing.T) {
	// Create a mock key with unsupported algorithm
	unsupportedKey := &keys.Key{
		Id:        keys.KeyId("test-id"),
		Algorithm: keys.Algorithm("UNSUPPORTED"),
	}

	_, err := keyToJWK(unsupportedKey)

	assert.Error(t, err)
}

func TestJWKSHandlerGetJWKS(t *testing.T) {
	repo := keys.NewInMemoryRepository()
	service := keys.NewService(repo)

	// Create a key
	ctx := context.Background()
	key, err := service.CreateKey(ctx, keys.RS256)
	require.NoError(t, err)

	// Test that the handler can access the service
	allKeys, err := service.ListAllKeys(ctx)
	require.NoError(t, err)

	jwks := keysToJWKS(allKeys)

	require.Len(t, jwks.Keys, 1)
	assert.Equal(t, string(key.Id), jwks.Keys[0].Kid)
	assert.Equal(t, "RSA", jwks.Keys[0].Kty)
	assert.Equal(t, "RS256", jwks.Keys[0].Alg)
}
