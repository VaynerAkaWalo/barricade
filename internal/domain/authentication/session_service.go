package authentication

import (
	"barricade/internal/infrastructure/htp"
	"context"
	"log/slog"
	"net/http"
)

type SessionRepository interface {
	Save(context.Context, *Session) error
	FindById(context.Context, SessionId) (*Session, error)
	FindByIdentity(context.Context, IdentityId) (*Session, error)
}

type IdentityRepository interface {
	FindByName(context.Context, string) (*Identity, error)
}

type SessionService struct {
	SessionStore  SessionRepository
	IdentityStore IdentityRepository
}

func (s *SessionService) AuthenticateBySession(ctx context.Context, sessionId SessionId) (*Session, error) {
	if sessionId == "" {
		return nil, htp.NewError("session id cannot be null or empty", http.StatusBadRequest)
	}

	session, err := s.SessionStore.FindById(ctx, sessionId)
	if err != nil {
		return nil, htp.NewError("session expired", http.StatusForbidden)
	}

	return session, err
}

func (s *SessionService) Login(ctx context.Context, name string, secret string) (*Session, error) {
	if name == "" || secret == "" {
		return nil, htp.NewError("name and secret cannot be empty", http.StatusBadRequest)
	}

	identity, err := s.IdentityStore.FindByName(ctx, name)
	if err != nil {
		return nil, htp.NewError("identity not found", http.StatusNotFound)
	}

	if err = identity.ValidateSecret(secret); err != nil {
		return nil, htp.NewError("invalid secret", http.StatusForbidden)
	}

	session, err := s.SessionStore.FindByIdentity(ctx, identity.Id)
	if err == nil {
		return session, nil
	}
	slog.ErrorContext(ctx, err.Error())

	session, err = NewSession(identity.Id)
	if err != nil {
		return nil, err
	}

	err = s.SessionStore.Save(ctx, session)
	if err != nil {
		return nil, htp.NewError("unable to create new session", http.StatusInternalServerError)
	}

	return session, nil
}
