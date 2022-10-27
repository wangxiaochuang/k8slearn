package uuid

import (
	"github.com/google/uuid"

	"k8s.io/apimachinery/pkg/types"
)

func NewUUID() types.UID {
	return types.UID(uuid.New().String())
}
