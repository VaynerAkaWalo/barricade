package authentication

import (
	"barricade/internal/infrastructure/htp"
	"barricade/pkg/uuid"
	"net/http"
	"time"
)

type SessionId string

type Session struct {
	Id        SessionId
	Owner     IdentityId
	CreatedAt int64
	ExpireAt  int64
}

func NewSession(owner IdentityId) (*Session, error) {
	if owner == "" {
		return nil, htp.NewError("session owner is required", http.StatusBadRequest)
	}

	createdAt := time.Now()

	return &Session{
		Id:        SessionId(uuid.New()),
		Owner:     owner,
		CreatedAt: createdAt.UnixMilli(),
		ExpireAt:  createdAt.Add(time.Minute * 5).Unix(),
	}, nil
}
