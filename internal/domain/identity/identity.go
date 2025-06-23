package identity

type Id string

type Identity struct {
	Id         Id
	Name       string
	SecretHash []byte
	CreatedAt  int64
	UpdatedAt  int64
}
