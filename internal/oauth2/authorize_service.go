package oauth2

import (
	"barricade/internal/identity"
	"barricade/internal/keys"
	"barricade/internal/oidc"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
)

type (
	responseType string
	scope        string
)

const (
	ResponseTypeIdToken responseType = "id_token"
	ResponseTypeCode    responseType = "code"
	ScopeOpenID         scope        = "openid"
)

type IdentityRepository interface {
	FindById(ctx context.Context, id identity.Id) (*identity.Identity, error)
}

type (
	AuthorizationParams struct {
		ResponseType        string
		ClientId            string
		Scope               string
		RedirectURI         string
		State               string
		CodeChallenge       string
		CodeChallengeMethod string
	}

	AuthorizationResult struct {
		IDToken oidc.IdToken
	}

	AuthorizeService struct {
		IdentityStore IdentityRepository
		ClientStore   ClientRepository
		CodeStore     AuthorizationCodeRepository
		KeyService    *keys.Service
		Issuer        string
		TokenExpiry   int
		CodeExpiry    int
	}
)

func (s *AuthorizeService) ValidateClientRedirect(ctx context.Context, params AuthorizationParams) (*Client, string, error) {
	client, err := s.ClientStore.FindById(ctx, ClientId(params.ClientId))
	if err != nil {
		if errors.Is(err, ErrClientNotFound) {
			return nil, "", ErrUnauthorizedClient
		}
		return nil, "", err
	}

	redirectURI := params.RedirectURI
	if redirectURI == "" {
		redirectURI = client.RedirectURI
	}

	parsedURI, err := url.ParseRequestURI(redirectURI)
	if err != nil {
		return nil, "", ErrInvalidRedirectURI
	}

	if !isRedirectDomainMatch(parsedURI.Hostname(), client.Domain) {
		return nil, "", ErrRedirectURIMismatch
	}

	return client, redirectURI, nil
}

func (s *AuthorizeService) Validate(params AuthorizationParams) error {
	if params.ResponseType == "" {
		return ErrInvalidRequest
	}

	if !strings.Contains(params.ResponseType, string(ResponseTypeIdToken)) && params.ResponseType != string(ResponseTypeCode) {
		return ErrUnsupportedResponseType
	}

	if params.ClientId == "" {
		return ErrInvalidRequest
	}

	if !strings.Contains(params.Scope, string(ScopeOpenID)) {
		return ErrInvalidScope
	}

	if params.CodeChallenge != "" {
		method := params.CodeChallengeMethod
		if method == "" {
			method = "S256"
		}
		if method != "S256" {
			return ErrInvalidCodeChallengeMethod
		}
		if !isValidCodeChallenge(params.CodeChallenge) {
			return ErrInvalidCodeChallenge
		}
	}

	return nil
}

func isValidCodeChallenge(challenge string) bool {
	l := len(challenge)
	return l >= 43 && l <= 128
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

func (s *AuthorizeService) GenerateCode(ctx context.Context, identityId identity.Id, params AuthorizationParams) (string, error) {
	code, err := generateAuthCode()
	if err != nil {
		return "", ErrServerError
	}

	method := params.CodeChallengeMethod
	if params.CodeChallenge != "" && method == "" {
		method = "S256"
	}

	authCode := NewAuthorizationCode(params.ClientId, string(identityId), params.RedirectURI, params.Scope, s.CodeExpiry)
	authCode.Code = code
	authCode.CodeChallenge = params.CodeChallenge
	authCode.CodeChallengeMethod = method

	err = s.CodeStore.Save(ctx, authCode)
	if err != nil {
		return "", ErrServerError
	}

	return code, nil
}

func generateAuthCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
