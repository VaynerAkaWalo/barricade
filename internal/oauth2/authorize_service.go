package oauth2

import (
	"barricade/internal/identity"
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
	ResponseTypeCode responseType = "code"
	ScopeOpenID      scope        = "openid"
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

	AuthorizeService struct {
		ClientStore ClientRepository
		CodeStore   AuthorizationCodeRepository
		CodeExpiry  int
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

	if params.ResponseType != string(ResponseTypeCode) {
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
