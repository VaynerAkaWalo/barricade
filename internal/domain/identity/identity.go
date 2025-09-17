package identity

import (
	"barricade/pkg/uuid"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
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
	if name == "" || secret == "" {
		return nil, xhttp.NewError("name and secret cannot be null or empty", http.StatusBadRequest)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(secret), 14)
	if err != nil {
		return nil, xhttp.NewError("internal error while creating secret hash", http.StatusInternalServerError)
	}

	createdAt := time.Now().UnixMilli()

	return &Identity{
		Id:         Id("ID_" + uuid.TrimmedUUID(16)),
		Name:       name,
		SecretHash: hash,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}, nil
}
