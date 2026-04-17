package oauth2

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type (
	ClientId     string
	ClientSecret string

	Client struct {
		Id          ClientId
		Name        string
		Domain      string
		SecretHash  []byte
		RedirectURI string
		CreatedAt   int64
		UpdatedAt   int64
	}
)

func NewClient(name string, domain string, redirectURI string) (*Client, ClientSecret, error) {
	if name == "" {
		return nil, "", ErrClientEmptyName
	}
	if domain == "" {
		return nil, "", ErrClientEmptyDomain
	}
	if redirectURI == "" {
		return nil, "", ErrClientEmptyRedirectURI
	}

	parsedURI, err := url.ParseRequestURI(redirectURI)
	if err != nil {
		return nil, "", ErrClientInvalidRedirectURI
	}

	if !isRedirectDomainMatch(parsedURI.Hostname(), domain) {
		return nil, "", ErrClientRedirectURIDomainMismatch
	}

	rawSecret, err := generateClientSecret()
	if err != nil {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(rawSecret), 14)
	if err != nil {
		return nil, "", err
	}

	createdAt := time.Now().UnixMilli()

	return &Client{
		Id:          ClientId(uuid.Must(uuid.NewV7()).String()),
		Name:        name,
		Domain:      domain,
		SecretHash:  hash,
		RedirectURI: redirectURI,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}, ClientSecret(rawSecret), nil
}

func generateClientSecret() (ClientSecret, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return ClientSecret(hex.EncodeToString(b)), nil
}

func isRedirectDomainMatch(host string, domain string) bool {
	return host == domain || strings.HasSuffix(host, "."+domain)
}
