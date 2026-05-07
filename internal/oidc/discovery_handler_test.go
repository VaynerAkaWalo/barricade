package oidc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryResponseStructure(t *testing.T) {
	handler := DiscoveryHandler{
		Issuer: "https://auth.example.com",
	}

	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()

	err := handler.GetDiscovery(w, req)
	require.NoError(t, err)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var body DiscoveryResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "https://auth.example.com", body.Issuer)
	assert.Equal(t, "https://auth.example.com/v1/oauth2/authorize", body.AuthorizationEndpoint)
	assert.Equal(t, "https://auth.example.com/v1/oauth2/token", body.TokenEndpoint)
	assert.Equal(t, "https://auth.example.com/v1/oauth2/userinfo", body.UserinfoEndpoint)
	assert.Equal(t, "https://auth.example.com/.well-known/jwks.json", body.JWKSUri)

	assert.Equal(t, []string{"openid", "profile"}, body.ScopesSupported)
	assert.Equal(t, []string{"code"}, body.ResponseTypesSupported)
	assert.Equal(t, []string{"query", "fragment"}, body.ResponseModesSupported)
	assert.Equal(t, []string{"public"}, body.SubjectTypesSupported)
	assert.Equal(t, []string{"RS256"}, body.IDTokenSigningAlgValuesSupported)
	assert.Equal(t, []string{"S256"}, body.CodeChallengeMethodsSupported)
	assert.Contains(t, body.ClaimsSupported, "sub")
	assert.Contains(t, body.ClaimsSupported, "auth_time")
	assert.Contains(t, body.ClaimsSupported, "acr")
	assert.Contains(t, body.ClaimsSupported, "amr")
	assert.Contains(t, body.ClaimsSupported, "nonce")
}


