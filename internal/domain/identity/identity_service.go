package identity

import (
	"context"
)

type Repository interface {
	Save(ctx context.Context, identity *Identity) error
	FindById(ctx context.Context, id Id) (*Identity, error)
}

type Service struct {
	Repo Repository
}

func (s *Service) Register(ctx context.Context, name string, secret string) (*Identity, error) {
	identity, err := New(name, secret)
	if err != nil {
		return nil, err
	}

	err = s.Repo.Save(ctx, identity)
	if err != nil {
		return nil, err
	}

	return identity, nil
}
