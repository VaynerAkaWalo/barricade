package identity

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Id string

type Identity struct {
	Id         Id
	Name       string
	SecretHash []byte
	CreatedAt  int64
	UpdatedAt  int64
}

func New(name string, secret string) (*Identity, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if secret == "" {
		return nil, ErrEmptySecret
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(secret), 14)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now().UnixMilli()

	return &Identity{
		Id:         Id(uuid.Must(uuid.NewV7()).String()),
		Name:       name,
		SecretHash: hash,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}, nil
}

func (i *Identity) ValidateSecret(secret string) error {
	return bcrypt.CompareHashAndPassword(i.SecretHash, []byte(secret))
}
