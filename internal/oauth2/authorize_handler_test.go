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

type mockClientRepository struct {
	findByIdFunc func(ctx context.Context, id ClientId) (*Client, error)
}

func (m *mockClientRepository) Save(ctx context.Context, client *Client) error {
	return nil
}

func (m *mockClientRepository) FindById(ctx context.Context, id ClientId) (*Client, error) {
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

func TestValidateClientRedirectClientNotFound(t *testing.T) {
	clientRepo := &mockClientRepository{
		findByIdFunc: func(ctx context.Context, id ClientId) (*Client, error) {
			return nil, ErrClientNotFound
		},
	}

	service := &AuthorizeService{ClientStore: clientRepo}

	params := AuthorizationParams{
		ClientId:    "unknown-client",
		RedirectURI: "https://example.com/callback",
	}

	_, _, err := service.ValidateClientRedirect(context.Background(), params)
	assert.ErrorIs(t, err, ErrUnauthorizedClient)
}

func TestValidateClientRedirectRedirectURIMismatch(t *testing.T) {
	clientRepo := &mockClientRepository{
		findByIdFunc: func(ctx context.Context, id ClientId) (*Client, error) {
			return &Client{Id: id, Domain: "example.com", RedirectURI: "https://example.com/callback"}, nil
		},
	}

	service := &AuthorizeService{ClientStore: clientRepo}

	params := AuthorizationParams{
		ClientId:    "test-client",
		RedirectURI: "https://evil.com/callback",
	}

	_, _, err := service.ValidateClientRedirect(context.Background(), params)
	assert.ErrorIs(t, err, ErrRedirectURIMismatch)
}

func TestValidateClientRedirectInvalidRedirectURI(t *testing.T) {
	clientRepo := &mockClientRepository{
		findByIdFunc: func(ctx context.Context, id ClientId) (*Client, error) {
			return &Client{Id: id, Domain: "example.com", RedirectURI: "https://example.com/callback"}, nil
		},
	}

	service := &AuthorizeService{ClientStore: clientRepo}

	params := AuthorizationParams{
		ClientId:    "test-client",
		RedirectURI: "://invalid-url",
	}

	_, _, err := service.ValidateClientRedirect(context.Background(), params)
	assert.ErrorIs(t, err, ErrInvalidRedirectURI)
}

func TestValidateClientRedirectUsesClientRedirectURIFallback(t *testing.T) {
	clientRepo := &mockClientRepository{
		findByIdFunc: func(ctx context.Context, id ClientId) (*Client, error) {
			return &Client{Id: id, Domain: "example.com", RedirectURI: "https://example.com/callback"}, nil
		},
	}

	service := &AuthorizeService{ClientStore: clientRepo}

	params := AuthorizationParams{
		ClientId:    "test-client",
		RedirectURI: "",
	}

	client, redirectURI, err := service.ValidateClientRedirect(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, "test-client", string(client.Id))
	assert.Equal(t, "https://example.com/callback", redirectURI)
}

func TestValidateClientRedirectSubdomainAllowed(t *testing.T) {
	clientRepo := &mockClientRepository{
		findByIdFunc: func(ctx context.Context, id ClientId) (*Client, error) {
			return &Client{Id: id, Domain: "example.com", RedirectURI: "https://example.com/callback"}, nil
		},
	}

	service := &AuthorizeService{ClientStore: clientRepo}

	params := AuthorizationParams{
		ClientId:    "test-client",
		RedirectURI: "https://sub.example.com/callback",
	}

	client, redirectURI, err := service.ValidateClientRedirect(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, "test-client", string(client.Id))
	assert.Equal(t, "https://sub.example.com/callback", redirectURI)
}

func TestValidateClientRedirectHappyPath(t *testing.T) {
	clientRepo := &mockClientRepository{
		findByIdFunc: func(ctx context.Context, id ClientId) (*Client, error) {
			return &Client{Id: id, Domain: "example.com", RedirectURI: "https://example.com/callback"}, nil
		},
	}

	service := &AuthorizeService{ClientStore: clientRepo}

	params := AuthorizationParams{
		ClientId:    "test-client",
		RedirectURI: "https://example.com/callback",
	}

	client, redirectURI, err := service.ValidateClientRedirect(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, "test-client", string(client.Id))
	assert.Equal(t, "https://example.com/callback", redirectURI)
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
