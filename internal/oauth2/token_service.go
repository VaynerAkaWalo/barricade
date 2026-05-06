package oauth2

import (
	"barricade/internal/identity"
	"barricade/internal/keys"
	"barricade/internal/oidc"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type TokenService struct {
	IdentityStore      IdentityRepository
	ClientStore        ClientRepository
	CodeStore          AuthorizationCodeRepository
	KeyService         *keys.Service
	Issuer             string
	TokenExpiry        int
	AccessTokenExpiry  int
}

type ExchangeTokenParams struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientId     string
	ClientSecret string
	CodeVerifier string
}

type TokenResult struct {
	IDToken     string `json:"id_token"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *TokenService) Exchange(ctx context.Context, params ExchangeTokenParams) (*TokenResult, error) {
	if params.GrantType != "authorization_code" {
		return nil, ErrUnsupportedGrantType
	}

	client, err := s.ClientStore.FindById(ctx, ClientId(params.ClientId))
	if err != nil {
		if err == ErrClientNotFound {
			return nil, ErrInvalidClient
		}
		return nil, err
	}

	authCode, err := s.CodeStore.FindByCode(ctx, params.Code)
	if err != nil {
		return nil, err
	}

	if time.Now().Unix() > authCode.ExpireAt {
		_ = s.CodeStore.Delete(ctx, params.Code)
		return nil, ErrCodeExpired
	}

	if authCode.ClientId != params.ClientId {
		return nil, ErrCodeMismatch
	}

	if authCode.RedirectURI != params.RedirectURI {
		return nil, ErrCodeMismatch
	}

	if authCode.CodeChallenge != "" {
		if params.CodeVerifier == "" {
			return nil, ErrMissingCodeVerifier
		}
		if !verifyPKCE(params.CodeVerifier, authCode.CodeChallenge) {
			return nil, ErrInvalidCodeVerifier
		}
	} else {
		if err := bcrypt.CompareHashAndPassword(client.SecretHash, []byte(params.ClientSecret)); err != nil {
			return nil, ErrInvalidClient
		}
	}

	if err := s.CodeStore.Delete(ctx, params.Code); err != nil {
		return nil, err
	}

	ident, err := s.IdentityStore.FindById(ctx, identity.Id(authCode.IdentityId))
	if err != nil {
		return nil, err
	}

	key, err := s.KeyService.GetSigningKey(ctx, keys.RS256)
	if err != nil {
		return nil, ErrServerError
	}

	privateKey, err := key.RSAPrivateKey()
	if err != nil {
		return nil, ErrServerError
	}

	idToken, err := oidc.NewIdToken(oidc.IdTokenParams{
		Key:           key,
		Ident:         ident,
		ClientId:      params.ClientId,
		Issuer:        s.Issuer,
		ExpiryMinutes: s.TokenExpiry,
	})
	if err != nil {
		return nil, ErrServerError
	}

	now := time.Now()
	accessExp := now.Add(time.Duration(s.AccessTokenExpiry) * time.Minute)

	accessClaims := jwt.MapClaims{
		"iss":       s.Issuer,
		"sub":       string(ident.Id),
		"aud":       s.Issuer,
		"exp":       accessExp.Unix(),
		"iat":       now.Unix(),
		"client_id": params.ClientId,
		"scope":     authCode.Scope,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessToken.Header["kid"] = string(key.Id)

	accessSigned, err := accessToken.SignedString(privateKey)
	if err != nil {
		return nil, ErrServerError
	}

	return &TokenResult{
		IDToken:     string(idToken),
		AccessToken: accessSigned,
		TokenType:   "Bearer",
		ExpiresIn:   s.AccessTokenExpiry * 60,
	}, nil
}

func verifyPKCE(verifier string, challenge string) bool {
	hash := sha256.Sum256([]byte(verifier))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])
	return encoded == challenge
}
