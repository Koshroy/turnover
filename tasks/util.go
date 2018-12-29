package tasks

import (
	"bytes"

	"github.com/gofrs/uuid"
)

func uuidEqual(uuid1 uuid.UUID, uuid2 uuid.UUID) bool {
	return bytes.Equal(uuid1.Bytes(), uuid2.Bytes())
}
