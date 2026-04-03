package keys

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, key *Key) error
	FindById(ctx context.Context, id KeyId) (*Key, error)
	FindLatest(ctx context.Context, algorithm Algorithm) (*Key, error)
	FindAll(ctx context.Context) ([]*Key, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateKey(ctx context.Context, algorithm Algorithm) (*Key, error) {
	var key *Key
	var err error

	switch algorithm {
	case RS256:
		key, err = s.createRS256Key()
	default:
		return nil, ErrUnsupportedAlgorithm
	}

	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, key); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *Service) createRS256Key() (*Key, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	publicKeyASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyASN1,
	})

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	createdAt := time.Now().UnixMilli()

	return &Key{
		Id:         KeyId(uuid.Must(uuid.NewV7()).String()),
		Algorithm:  RS256,
		CreatedAt:  createdAt,
		PublicKey:  publicKeyPEM,
		PrivateKey: privateKeyPEM,
	}, nil
}

func (s *Service) GetKey(ctx context.Context, id KeyId) (*Key, error) {
	return s.repo.FindById(ctx, id)
}

func (s *Service) ListAllKeys(ctx context.Context) ([]*Key, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetSigningKey(ctx context.Context, algorithm Algorithm) (*Key, error) {
	return s.repo.FindLatest(ctx, algorithm)
}
