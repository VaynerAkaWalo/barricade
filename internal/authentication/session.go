package authentication

import (
	"barricade/internal/identity"
	"barricade/pkg/uuid"
	"time"
)

type SessionId string

type Session struct {
	Id        SessionId
	Owner     identity.Id
	CreatedAt int64
	ExpireAt  int64
}

func NewSession(owner identity.Id) (*Session, error) {
	if owner == "" {
		return nil, ErrInvalidSessionOwner
	}

	createdAt := time.Now()

	return &Session{
		Id:        SessionId(uuid.New()),
		Owner:     owner,
		CreatedAt: createdAt.UnixMilli(),
		ExpireAt:  createdAt.Add(time.Minute * 5).Unix(),
	}, nil
}
