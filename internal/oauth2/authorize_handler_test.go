package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
