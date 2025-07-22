package authentication

import "golang.org/x/crypto/bcrypt"

type IdentityId string

type Identity struct {
	Id         IdentityId
	Name       string
	SecretHash []byte
}

func (i *Identity) ValidateSecret(secret string) error {
	return bcrypt.CompareHashAndPassword(i.SecretHash, []byte(secret))
}
