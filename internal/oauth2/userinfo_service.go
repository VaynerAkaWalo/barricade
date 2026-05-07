package oauth2

import (
	"barricade/internal/identity"
	"barricade/internal/keys"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type UserinfoResult struct {
	Sub  string
	Name string
}

type UserinfoService struct {
	KeyService    *keys.Service
	IdentityStore IdentityRepository
	Issuer        string
}

func (s *UserinfoService) GetUserinfo(ctx context.Context, accessToken string) (*UserinfoResult, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid")
		}

		key, err := s.KeyService.GetKey(ctx, keys.KeyId(kid))
		if err != nil {
			return nil, fmt.Errorf("key not found: %w", err)
		}

		return key.RSAPublicKey()
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return nil, ErrInvalidToken
	}

	issuer, _ := claims["iss"].(string)
	if issuer != s.Issuer {
		return nil, ErrInvalidToken
	}

	scope, _ := claims["scope"].(string)

	if !strings.Contains(scope, "openid") {
		return nil, ErrInsufficientScope
	}

	ident, err := s.IdentityStore.FindById(ctx, identity.Id(sub))
	if err != nil {
		return nil, ErrInvalidToken
	}

	result := &UserinfoResult{
		Sub: string(ident.Id),
	}

	if strings.Contains(scope, "profile") {
		result.Name = ident.Name
	}

	return result, nil
}

func mapUserinfoError(err error) (string, int) {
	switch {
	case errors.Is(err, ErrInvalidToken):
		return "invalid_token", 401
	case errors.Is(err, ErrInsufficientScope):
		return "insufficient_scope", 403
	default:
		return "server_error", 500
	}
}

func userinfoWWWAuthenticate(code string) string {
	switch code {
	case "invalid_token":
		return `Bearer error="invalid_token"`
	case "insufficient_scope":
		return `Bearer error="insufficient_scope", scope="openid"`
	default:
		return ""
	}
}
