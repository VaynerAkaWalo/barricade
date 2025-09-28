package uuid

import (
	"github.com/google/uuid"
)

func New() string {
	val, _ := uuid.NewV7()
	return val.String()
}
