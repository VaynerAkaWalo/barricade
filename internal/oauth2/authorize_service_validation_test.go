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
		ResponseType: "id_token",
		ClientId:     "test-client",
		Scope:        "profile email",
	})

	assert.ErrorIs(t, err, ErrInvalidScope)
}

func TestAuthorizeServiceValidateHappyPath(t *testing.T) {
	svc := &AuthorizeService{}

	err := svc.Validate(AuthorizationParams{
		ResponseType: "id_token",
		ClientId:     "test-client",
		Scope:        "openid profile",
	})

	assert.NoError(t, err)
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
