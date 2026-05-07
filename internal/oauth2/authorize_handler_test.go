package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildErrorRedirectURL(t *testing.T) {
	redirectURL := buildErrorRedirectURL("https://example.com/callback", "invalid_request", "missing parameter")

	assert.Contains(t, redirectURL, "https://example.com/callback")
	assert.Contains(t, redirectURL, "error=invalid_request")
	assert.Contains(t, redirectURL, "error_description=missing+parameter")
}

func TestBuildErrorRedirectURLEmptyDescription(t *testing.T) {
	redirectURL := buildErrorRedirectURL("https://example.com/callback", "invalid_scope", "")

	assert.Contains(t, redirectURL, "error=invalid_scope")
	assert.NotContains(t, redirectURL, "error_description")
}

func TestBuildErrorRedirectURLInvalidURI(t *testing.T) {
	redirectURL := buildErrorRedirectURL("://invalid-url", "invalid_request", "missing parameter")

	assert.Equal(t, "://invalid-url", redirectURL)
}
