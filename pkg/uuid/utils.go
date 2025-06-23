package uuid

import (
	"github.com/google/uuid"
	"strings"
)

func TrimmedUUID(length int) string {
	val, _ := uuid.NewV7()
	return strings.ReplaceAll(val.String(), "-", "")[:length]
}
