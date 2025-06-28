package identity

import (
	"barricade/internal/infrastructure/htp"
	"barricade/pkg/uuid"
	"context"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"time"
)

type Repository interface {
	Save(ctx context.Context, identity *Identity) error
	FindById(ctx context.Context, id Id) (*Identity, error)
}

type Service struct {
	Repo Repository
}

func (s *Service) Register(ctx context.Context, name string, secret string) (*Identity, error) {
	if name == "" || secret == "" {
		return nil, htp.NewError("name and secret cannot be null or empty", http.StatusBadRequest)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(secret), 14)

	createdAt := time.Now().UnixMilli()

	identity := &Identity{
		Id:         Id("ID_" + uuid.TrimmedUUID(16)),
		Name:       name,
		SecretHash: hash,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}

	err = s.Repo.Save(ctx, identity)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, err
	}

	return identity, nil
}
