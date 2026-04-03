package keys

import "errors"

var (
	ErrNoSigningKey         = errors.New("no signing key available")
	ErrKeyNotFound          = errors.New("key not found")
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrInvalidKey           = errors.New("invalid key")
)
