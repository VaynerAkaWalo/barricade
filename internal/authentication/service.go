package authentication

import (
	"barricade/internal/identity"
	"context"
	"net/http"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type Service struct {
	SessionStore  SessionRepository
	IdentityStore IdentityRepository
}

func (service *Service) AuthenticateBySession(ctx context.Context, sessionId SessionId) (*identity.Identity, error) {
	if sessionId == "" {
		return nil, xhttp.NewError("session id cannot be null or empty", http.StatusBadRequest)
	}

	session, err := service.SessionStore.FindById(ctx, sessionId)
	if err != nil {
		return nil, xhttp.NewError("session expired", http.StatusForbidden)
	}

	identity, err := service.IdentityStore.FindById(ctx, session.Owner)
	if err != nil {
		return nil, xhttp.NewError("identity for session not found", http.StatusForbidden)
	}

	return identity, nil
}
