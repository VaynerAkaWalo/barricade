package authentication

import (
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"net/http"
)

type SessionRepository interface {
	Save(context.Context, *Session) error
	FindById(context.Context, SessionId) (*Session, error)
	FindByIdentity(context.Context, IdentityId) (*Session, error)
}

type IdentityRepository interface {
	FindByName(context.Context, string) (*Identity, error)
	FindById(context.Context, IdentityId) (*Identity, error)
}

type SessionService struct {
	SessionStore  SessionRepository
	IdentityStore IdentityRepository
}

func (s *SessionService) GetIdentityBySession(ctx context.Context, sessionId SessionId) (*Identity, error) {
	if sessionId == "" {
		return nil, xhttp.NewError("session id cannot be null or empty", http.StatusBadRequest)
	}

	session, err := s.SessionStore.FindById(ctx, sessionId)
	if err != nil {
		return nil, xhttp.NewError("session expired", http.StatusForbidden)
	}

	identity, err := s.IdentityStore.FindById(ctx, session.Owner)
	if err != nil {
		return nil, xhttp.NewError("identity for session not found", http.StatusForbidden)
	}

	return identity, nil
}

func (s *SessionService) Login(ctx context.Context, name string, secret string) (*Session, error) {
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
