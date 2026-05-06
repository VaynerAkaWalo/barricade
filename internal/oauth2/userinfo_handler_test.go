package oauth2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"barricade/internal/identity"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func setupUserinfoTest(t *testing.T) (*oauth2Module, *UserinfoHandler, identity.Identity) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	handler := &UserinfoHandler{
		KeyService:    module.keyService,
		IdentityStore: module.identityRepository,
		Issuer:        "https://test.issuer.com",
	}

	return module, handler, *ident
}

func signTestAccessToken(t *testing.T, module *oauth2Module, claims jwt.MapClaims) string {
	key, err := module.keyService.GetSigningKey(context.Background(), "RS256")
	assert.NoError(t, err)

	privateKey, err := key.RSAPrivateKey()
	assert.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = string(key.Id)

	signed, err := token.SignedString(privateKey)
	assert.NoError(t, err)

	return signed
}

func TestUserinfoHappyPath(t *testing.T) {
	module, handler, ident := setupUserinfoTest(t)

	token := signTestAccessToken(t, module, jwt.MapClaims{
		"iss":       "https://test.issuer.com",
		"sub":       string(ident.Id),
		"aud":       "https://test.issuer.com",
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
		"client_id": "test-client",
		"scope":     "openid profile",
	})

	req := httptest.NewRequest("GET", "/v1/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	err := handler.Userinfo(rec, req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp userinfoResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, string(ident.Id), resp.Sub)
	assert.Equal(t, ident.Name, resp.Name)
}

func TestUserinfoHappyPathWithoutProfileScope(t *testing.T) {
	module, handler, ident := setupUserinfoTest(t)

	token := signTestAccessToken(t, module, jwt.MapClaims{
		"iss":       "https://test.issuer.com",
		"sub":       string(ident.Id),
		"aud":       "https://test.issuer.com",
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
		"client_id": "test-client",
		"scope":     "openid",
	})

	req := httptest.NewRequest("GET", "/v1/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	err := handler.Userinfo(rec, req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp userinfoResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, string(ident.Id), resp.Sub)
	assert.Empty(t, resp.Name)
}

func TestUserinfoMissingToken(t *testing.T) {
	_, handler, _ := setupUserinfoTest(t)

	req := httptest.NewRequest("GET", "/v1/oauth2/userinfo", nil)
	rec := httptest.NewRecorder()

	err := handler.Userinfo(rec, req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, err.(*xhttp.HttpError).Code)
	assert.Contains(t, rec.Header().Get("WWW-Authenticate"), "invalid_token")
}

func TestUserinfoEmptyBearer(t *testing.T) {
	_, handler, _ := setupUserinfoTest(t)

	req := httptest.NewRequest("GET", "/v1/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()

	err := handler.Userinfo(rec, req)
	assert.Error(t, err)
}

func TestUserinfoInvalidToken(t *testing.T) {
	_, handler, _ := setupUserinfoTest(t)

	req := httptest.NewRequest("GET", "/v1/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer obviously-invalid-token")
	rec := httptest.NewRecorder()

	err := handler.Userinfo(rec, req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, err.(*xhttp.HttpError).Code)
}

func TestUserinfoExpiredToken(t *testing.T) {
	module, handler, ident := setupUserinfoTest(t)

	token := signTestAccessToken(t, module, jwt.MapClaims{
		"iss":       "https://test.issuer.com",
		"sub":       string(ident.Id),
		"aud":       "https://test.issuer.com",
		"exp":       time.Now().Add(-1 * time.Hour).Unix(),
		"iat":       time.Now().Add(-2 * time.Hour).Unix(),
		"client_id": "test-client",
		"scope":     "openid",
	})

	req := httptest.NewRequest("GET", "/v1/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	err := handler.Userinfo(rec, req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, err.(*xhttp.HttpError).Code)
}

func TestUserinfoWrongIssuer(t *testing.T) {
	module, handler, ident := setupUserinfoTest(t)

	token := signTestAccessToken(t, module, jwt.MapClaims{
		"iss":       "https://evil-issuer.com",
		"sub":       string(ident.Id),
		"aud":       "https://evil-issuer.com",
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
		"client_id": "test-client",
		"scope":     "openid",
	})

	req := httptest.NewRequest("GET", "/v1/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	err := handler.Userinfo(rec, req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, err.(*xhttp.HttpError).Code)
}

func TestUserinfoMissingOpenidScope(t *testing.T) {
	module, handler, ident := setupUserinfoTest(t)

	token := signTestAccessToken(t, module, jwt.MapClaims{
		"iss":       "https://test.issuer.com",
		"sub":       string(ident.Id),
		"aud":       "https://test.issuer.com",
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
		"client_id": "test-client",
		"scope":     "profile",
	})

	req := httptest.NewRequest("GET", "/v1/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	err := handler.Userinfo(rec, req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, err.(*xhttp.HttpError).Code)
	assert.Contains(t, rec.Header().Get("WWW-Authenticate"), "insufficient_scope")
}
