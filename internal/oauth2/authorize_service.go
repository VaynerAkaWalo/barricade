package oauth2

import (
	"barricade/internal/identity"
	"barricade/internal/keys"
	"barricade/internal/oidc"
	"context"
	"strings"
)

type (
	responseType string
	scope        string
)

const (
	ResponseTypeIdToken responseType = "id_token"
	ScopeOpenID         scope        = "openid"
)

type IdentityRepository interface {
	FindById(ctx context.Context, id identity.Id) (*identity.Identity, error)
}

type (
	AuthorizationParams struct {
		ResponseType string
		ClientId     string
		Scope        string
	}

	AuthorizationResult struct {
		IDToken oidc.IdToken
	}

	AuthorizeService struct {
		IdentityStore IdentityRepository
		KeyService    *keys.Service
		Issuer        string
		TokenExpiry   int
	}
)

func (s *AuthorizeService) Validate(params AuthorizationParams) error {
	if params.ResponseType == "" {
		return ErrInvalidRequest
	}

	if !strings.Contains(params.ResponseType, string(ResponseTypeIdToken)) {
		return ErrUnsupportedResponseType
	}

	if params.ClientId == "" {
		return ErrInvalidRequest
	}

	if !strings.Contains(params.Scope, string(ScopeOpenID)) {
		return ErrInvalidScope
	}

	return nil
}

func (s *AuthorizeService) Authorize(ctx context.Context, identityId identity.Id, clientId string) (*AuthorizationResult, error) {
	ident, err := s.IdentityStore.FindById(ctx, identityId)
	if err != nil {
		return nil, err
	}

	key, err := s.KeyService.GetSigningKey(ctx, keys.RS256)
	if err != nil {
		return nil, ErrServerError
	}

	token, err := oidc.NewIdToken(oidc.IdTokenParams{
		Key:           key,
		Ident:         ident,
		ClientId:      clientId,
		Issuer:        s.Issuer,
		ExpiryMinutes: s.TokenExpiry,
	})
	if err != nil {
		return nil, ErrServerError
	}

	return &AuthorizationResult{
		IDToken: token,
	}, nil
}
