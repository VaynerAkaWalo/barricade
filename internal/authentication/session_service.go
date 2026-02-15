package authentication

import (
	"barricade/internal/identity"
	"context"
	"net/http"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type SessionRepository interface {
	Save(context.Context, *Session) error
	FindById(context.Context, SessionId) (*Session, error)
	FindByIdentity(context.Context, identity.Id) (*Session, error)
}

type IdentityRepository interface {
	FindByName(context.Context, string) (*identity.Identity, error)
	FindById(context.Context, identity.Id) (*identity.Identity, error)
}

type SessionService struct {
	SessionStore  SessionRepository
	IdentityStore IdentityRepository
}

func (s *SessionService) CreateOrGetSessionForCredentials(ctx context.Context, name string, secret string) (*Session, error) {
	if name == "" || secret == "" {
		return nil, xhttp.NewError("name and secret cannot be empty", http.StatusBadRequest)
	}

	identity, err := s.IdentityStore.FindByName(ctx, name)
	if err != nil {
		return nil, xhttp.NewError("identity not found", http.StatusNotFound)
	}

	if err = identity.ValidateSecret(secret); err != nil {
		return nil, xhttp.NewError("invalid secret", http.StatusForbidden)
	}

	session, err := s.SessionStore.FindByIdentity(ctx, identity.Id)
	if err == nil {
		return session, nil
	}

	session, err = NewSession(identity.Id)
	if err != nil {
		return nil, err
	}

	err = s.SessionStore.Save(ctx, session)
	if err != nil {
		return nil, xhttp.NewError("unable to create new session", http.StatusInternalServerError)
	}

	return session, nil
}
