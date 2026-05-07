package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizeServiceValidateMissingResponseType(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ClientId: "test-client",
		Scope:    "openid",
	})

	assert.ErrorIs(t, err, ErrInvalidRequest)
}

func TestAuthorizeServiceValidateUnsupportedResponseType(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ResponseType: "token",
		ClientId:     "test-client",
		Scope:        "openid",
	})

	assert.ErrorIs(t, err, ErrUnsupportedResponseType)
}

func TestAuthorizeServiceValidateMissingOpenIDScope(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ResponseType: "code",
		ClientId:     "test-client",
		Scope:        "profile email",
	})

	assert.ErrorIs(t, err, ErrInvalidScope)
}

func TestAuthorizeServiceValidateCodeResponseType(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ResponseType: "code",
		ClientId:     "test-client",
		Scope:        "openid",
	})

	assert.NoError(t, err)
}

func TestAuthorizeServiceValidatePKCEHappyPath(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ResponseType:        "code",
		ClientId:            "test-client",
		Scope:               "openid",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
	})

	assert.NoError(t, err)
}

func TestAuthorizeServiceValidatePKCEDefaultsToS256(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ResponseType:  "code",
		ClientId:      "test-client",
		Scope:         "openid",
		CodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
	})

	assert.NoError(t, err)
}

func TestAuthorizeServiceValidatePKCEInvalidMethod(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ResponseType:        "code",
		ClientId:            "test-client",
		Scope:               "openid",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "plain",
	})

	assert.ErrorIs(t, err, ErrInvalidCodeChallengeMethod)
}

func TestAuthorizeServiceValidatePKCEChallengeTooShort(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ResponseType:        "code",
		ClientId:            "test-client",
		Scope:               "openid",
		CodeChallenge:       "short",
		CodeChallengeMethod: "S256",
	})

	assert.ErrorIs(t, err, ErrInvalidCodeChallenge)
}

func TestAuthorizeServiceValidatePKCEChallengeTooLong(t *testing.T) {
	svc := &AuthorizeService{}

	challenge := make([]byte, 129)
	for i := range challenge {
		challenge[i] = 'a'
	}

	err := svc.Validate(AuthorizationParams{
		ResponseType:        "code",
		ClientId:            "test-client",
		Scope:               "openid",
		CodeChallenge:       string(challenge),
		CodeChallengeMethod: "S256",
	})

	assert.ErrorIs(t, err, ErrInvalidCodeChallenge)
}


