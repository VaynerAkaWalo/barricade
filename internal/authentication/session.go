package authentication

import (
	"time"

	"barricade/internal/identity"

	"github.com/google/uuid"
)

type (
	SessionId string
	Session   struct {
		Id        SessionId
		Owner     identity.Id
		CreatedAt int64
		ExpireAt  int64
	}
)

func NewSession(owner identity.Id) (*Session, error) {
	if owner == "" {
		return nil, ErrInvalidSessionOwner
	}

	createdAt := time.Now()

	return &Session{
		Id:        SessionId(uuid.Must(uuid.NewV7()).String()),
		Owner:     owner,
		CreatedAt: createdAt.UnixMilli(),
		ExpireAt:  createdAt.Add(time.Hour * 1).Unix(),
	}, nil
}
