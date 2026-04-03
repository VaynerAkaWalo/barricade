package keys

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

type KeyId string
type Algorithm string

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
		return k.signRS256(data)
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}

func (k *Key) signer() (crypto.Signer, error) {
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

	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, ErrInvalidKey
	}

	return signer, nil
}

func (k *Key) signRS256(data []byte) ([]byte, error) {
	signer, err := k.signer()
	if err != nil {
		return nil, err
	}

	hash := crypto.SHA256.New()
	hash.Write(data)
	digest := hash.Sum(nil)

	rsaPrivateKey, ok := signer.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidKey
	}

	return rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, digest)
}
