package authentication

import (
	"barricade/internal/identity"
	"context"
	"time"
)

type Service struct {
	SessionStore  SessionRepository
	IdentityStore IdentityRepository
}

func (service *Service) AuthenticateBySession(ctx context.Context, sessionId SessionId) (*identity.Identity, error) {
	if sessionId == "" {
		return nil, ErrEmptySessionId
	}

	session, err := service.SessionStore.FindById(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	if time.Now().Unix() > session.ExpireAt {
		return nil, ErrSessionExpired
	}

	ident, err := service.IdentityStore.FindById(ctx, session.Owner)
	if err != nil {
		return nil, err
	}

	return ident, nil
}
