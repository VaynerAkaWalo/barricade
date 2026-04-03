package keys

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateKeyRS256(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	key, err := service.CreateKey(ctx, RS256)

	require.NoError(t, err)
	assert.NotEmpty(t, key.Id)
	assert.Equal(t, RS256, key.Algorithm)
	assert.NotNil(t, key.PrivateKey)
	assert.NotNil(t, key.PublicKey)
	assert.Greater(t, key.CreatedAt, int64(0))
}

func TestCreateKeyUnsupportedAlgorithm(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	type UnsupportedAlgorithm Algorithm
	const UNSUPPORTED UnsupportedAlgorithm = "UNSUPPORTED"

	_, err := service.CreateKey(ctx, Algorithm(UNSUPPORTED))

	assert.ErrorIs(t, err, ErrUnsupportedAlgorithm)
}

func TestGetSigningKeyNoKey(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	_, err := service.GetSigningKey(ctx, RS256)

	assert.ErrorIs(t, err, ErrNoSigningKey)
}

func TestGetSigningKeyReturnsLatest(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	// Create first key
	key1, err := service.CreateKey(ctx, RS256)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Create second key (should be newer)
	key2, err := service.CreateKey(ctx, RS256)
	require.NoError(t, err)

	// Get signing key - should return the latest (key2)
	signingKey, err := service.GetSigningKey(ctx, RS256)

	require.NoError(t, err)
	assert.Equal(t, key2.Id, signingKey.Id)
	assert.Greater(t, signingKey.CreatedAt, key1.CreatedAt)
}

func TestGetKeyNotFound(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	_, err := service.GetKey(ctx, KeyId("nonexistent"))

	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestKeySignRS256(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	key, err := service.CreateKey(ctx, RS256)
	require.NoError(t, err)

	data := []byte("test data to sign")
	signature, err := key.Sign(data)

	require.NoError(t, err)
	assert.NotEmpty(t, signature)
}
