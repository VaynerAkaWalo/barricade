package oauth2

import (
	"barricade/internal/identity"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockIdentityRepository struct {
	findByIdFunc func(ctx context.Context, id identity.Id) (*identity.Identity, error)
}

func (m *mockIdentityRepository) FindById(ctx context.Context, id identity.Id) (*identity.Identity, error) {
	return m.findByIdFunc(ctx, id)
}

func TestAuthorizeServiceValidateMissingResponseType(t *testing.T) {
	service := &AuthorizeService{}

	params := AuthorizationParams{
		ClientId: "test-client",
		Scope:    "openid",
	}

	err := service.Validate(params)
	assert.ErrorIs(t, err, ErrInvalidRequest)
}

func TestAuthorizeServiceValidateUnsupportedResponseType(t *testing.T) {
	service := &AuthorizeService{}

	params := AuthorizationParams{
		ResponseType: "code",
		ClientId:     "test-client",
		Scope:        "openid",
	}

	err := service.Validate(params)
	assert.ErrorIs(t, err, ErrUnsupportedResponseType)
}

func TestAuthorizeServiceValidateMissingClientId(t *testing.T) {
	service := &AuthorizeService{}

	params := AuthorizationParams{
		ResponseType: "id_token",
		Scope:        "openid",
	}

	err := service.Validate(params)
	assert.ErrorIs(t, err, ErrInvalidRequest)
}

func TestAuthorizeServiceValidateMissingOpenIDScope(t *testing.T) {
	service := &AuthorizeService{}

	params := AuthorizationParams{
		ResponseType: "id_token",
		ClientId:     "test-client",
		Scope:        "profile email",
	}

	err := service.Validate(params)
	assert.ErrorIs(t, err, ErrInvalidScope)
}

func TestAuthorizeServiceValidateHappyPath(t *testing.T) {
	service := &AuthorizeService{}

	params := AuthorizationParams{
		ResponseType: "id_token",
		ClientId:     "test-client",
		Scope:        "openid profile",
	}

	err := service.Validate(params)
	assert.NoError(t, err)
}

func TestBuildSuccessRedirectURL(t *testing.T) {
	result := &AuthorizationResult{
		IDToken: "test-id-token",
	}

	redirectURL := buildSuccessRedirectURL("https://example.com/callback", result, 5)

	assert.Contains(t, redirectURL, "https://example.com/callback")
	assert.Contains(t, redirectURL, "id_token=test-id-token")
	assert.Contains(t, redirectURL, "token_type=Bearer")
	assert.Contains(t, redirectURL, "expires_in=300")
	assert.Contains(t, redirectURL, "#")
}

func TestBuildErrorRedirectURL(t *testing.T) {
	redirectURL := buildErrorRedirectURL("https://example.com/callback", "invalid_request", "missing parameter")

	assert.Contains(t, redirectURL, "https://example.com/callback")
	assert.Contains(t, redirectURL, "error=invalid_request")
	assert.Contains(t, redirectURL, "error_description=missing+parameter")
	assert.Contains(t, redirectURL, "#")
}

func TestBuildErrorRedirectURLEmptyDescription(t *testing.T) {
	redirectURL := buildErrorRedirectURL("https://example.com/callback", "invalid_scope", "")

	assert.Contains(t, redirectURL, "error=invalid_scope")
	assert.NotContains(t, redirectURL, "error_description")
}

func TestValidateRedirectURIEmpty(t *testing.T) {
	err := validateRedirectURI("")
	assert.ErrorIs(t, err, ErrInvalidRequest)
}

func TestValidateRedirectURIInvalid(t *testing.T) {
	err := validateRedirectURI("://invalid-url")
	assert.ErrorIs(t, err, ErrInvalidRedirectURI)
}

func TestBuildSuccessRedirectURLInvalidURI(t *testing.T) {
	result := &AuthorizationResult{
		IDToken: "test-id-token",
	}

	redirectURL := buildSuccessRedirectURL("://invalid-url", result, 5)

	assert.Equal(t, "://invalid-url", redirectURL)
}

func TestBuildErrorRedirectURLInvalidURI(t *testing.T) {
	redirectURL := buildErrorRedirectURL("://invalid-url", "invalid_request", "missing parameter")

	assert.Equal(t, "://invalid-url", redirectURL)
}
