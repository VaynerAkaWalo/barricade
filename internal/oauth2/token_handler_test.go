package oauth2

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/stretchr/testify/assert"
)

func TestTokenHandlerExchangeHappyPath(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	code := NewAuthorizationCode(string(clientResult.Client.Id), string(ident.Id), "https://example.com/callback", "openid", 5)
	code.Code = "handler-code-123"
	err = module.authCodeRepository.Save(context.Background(), code)
	assert.NoError(t, err)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "handler-code-123")
	form.Set("redirect_uri", "https://example.com/callback")
	form.Set("client_id", string(clientResult.Client.Id))
	form.Set("client_secret", string(clientResult.ClientSecret))

	req := httptest.NewRequest("POST", "/v1/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler := TokenHttpHandler{Service: module.tokenService}
	err = handler.Token(rec, req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "id_token")
	assert.Contains(t, rec.Body.String(), "Bearer")
}

func TestTokenHandlerInvalidClient(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "handler-code-123")
	form.Set("redirect_uri", "https://example.com/callback")
	form.Set("client_id", string(clientResult.Client.Id))
	form.Set("client_secret", "wrong-secret")

	req := httptest.NewRequest("POST", "/v1/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler := TokenHttpHandler{Service: module.tokenService}
	err = handler.Token(rec, req)
	assert.Error(t, err)

	xErr, ok := err.(*xhttp.HttpError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, xErr.Code)
}

func TestTokenHandlerUnsupportedGrantType(t *testing.T) {
	module := setupTokenModule(t)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req := httptest.NewRequest("POST", "/v1/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler := TokenHttpHandler{Service: module.tokenService}
	err := handler.Token(rec, req)
	assert.Error(t, err)

	xErr, ok := err.(*xhttp.HttpError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, xErr.Code)
}

func TestTokenHandlerExchangeWithBasicAuth(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	code := NewAuthorizationCode(string(clientResult.Client.Id), string(ident.Id), "https://example.com/callback", "openid", 5)
	code.Code = "basic-auth-code-123"
	err = module.authCodeRepository.Save(context.Background(), code)
	assert.NoError(t, err)

	credentials := base64.StdEncoding.EncodeToString([]byte(string(clientResult.Client.Id) + ":" + string(clientResult.ClientSecret)))

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "basic-auth-code-123")
	form.Set("redirect_uri", "https://example.com/callback")

	req := httptest.NewRequest("POST", "/v1/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+credentials)
	rec := httptest.NewRecorder()

	handler := TokenHttpHandler{Service: module.tokenService}
	err = handler.Token(rec, req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "id_token")
	assert.Contains(t, rec.Body.String(), "Bearer")
}

func TestTokenHandlerBasicAuthOverridesBody(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	code := NewAuthorizationCode(string(clientResult.Client.Id), string(ident.Id), "https://example.com/callback", "openid", 5)
	code.Code = "override-code-123"
	err = module.authCodeRepository.Save(context.Background(), code)
	assert.NoError(t, err)

	credentials := base64.StdEncoding.EncodeToString([]byte(string(clientResult.Client.Id) + ":" + string(clientResult.ClientSecret)))

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "override-code-123")
	form.Set("redirect_uri", "https://example.com/callback")
	form.Set("client_id", "wrong-client-id")
	form.Set("client_secret", "wrong-secret")

	req := httptest.NewRequest("POST", "/v1/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+credentials)
	rec := httptest.NewRecorder()

	handler := TokenHttpHandler{Service: module.tokenService}
	err = handler.Token(rec, req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "id_token")
}
