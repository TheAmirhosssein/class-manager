package base

import (
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/types"
)

type Entity struct {
	ID        types.ID  `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
