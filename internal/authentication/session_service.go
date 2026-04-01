package authentication

import (
	"barricade/internal/identity"
	"context"
	"errors"
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
	if name == "" {
		return nil, ErrEmptyName
	}
	if secret == "" {
		return nil, ErrEmptySecret
	}

	ident, err := s.IdentityStore.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, identity.ErrIdentityNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err = ident.ValidateSecret(secret); err != nil {
		return nil, ErrInvalidCredentials
	}

	session, err := s.SessionStore.FindByIdentity(ctx, ident.Id)
	if err == nil {
		return session, nil
	}
	if !errors.Is(err, ErrSessionNotFound) {
		return nil, err
	}

	session, err = NewSession(ident.Id)
	if err != nil {
		return nil, err
	}

	if err = s.SessionStore.Save(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}
