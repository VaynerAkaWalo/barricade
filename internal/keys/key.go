package keys

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

type (
	KeyId     string
	Algorithm string
)

const (
	RS256 Algorithm = "RS256"
)

type Key struct {
	Id         KeyId
	Algorithm  Algorithm
	CreatedAt  int64
	PublicKey  []byte
	PrivateKey []byte
}

func (k *Key) Sign(data []byte) ([]byte, error) {
	switch k.Algorithm {
	case RS256:
		privateKey, err := k.RSAPrivateKey()
		if err != nil {
			return nil, err
		}
		hash := crypto.SHA256.New()
		hash.Write(data)
		digest := hash.Sum(nil)
		return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest)
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}

func (k *Key) RSAPrivateKey() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(k.PrivateKey)
	if block == nil {
		return nil, ErrInvalidKey
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, ErrInvalidKey
		}
	}

	rsaPrivateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidKey
	}

	return rsaPrivateKey, nil
}
