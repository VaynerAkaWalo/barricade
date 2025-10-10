package ihttp

import (
	"barricade/internal/domain/authentication"
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"log/slog"
	"net/http"
)

type BarricadeAuthenticationProvider struct {
	SessionService authentication.SessionService
}

func (provider BarricadeAuthenticationProvider) FetchUser(ctx context.Context, token string, schema string) (xhttp.User, error) {
	if schema != xhttp.SessionV1 {
		slog.ErrorContext(ctx, "unsupported schema type: "+schema)
		return xhttp.User{}, xhttp.NewError("unsupported schema type", http.StatusBadRequest)
	}

	identity, err := provider.SessionService.GetIdentityBySession(ctx, authentication.SessionId(token))
	if err != nil {
		return xhttp.User{}, err
	}
	return xhttp.User{
		UserId: string(identity.Id),
	}, nil
}
