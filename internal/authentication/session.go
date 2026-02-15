package authentication

import (
	"barricade/internal/identity"
	"barricade/pkg/uuid"
	"net/http"
	"time"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
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
		return nil, xhttp.NewError("session owner is required", http.StatusBadRequest)
	}

	createdAt := time.Now()

	return &Session{
		Id:        SessionId(uuid.New()),
		Owner:     owner,
		CreatedAt: createdAt.UnixMilli(),
		ExpireAt:  createdAt.Add(time.Minute * 5).Unix(),
	}, nil
}
