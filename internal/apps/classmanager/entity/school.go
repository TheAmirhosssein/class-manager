package entity

import "github.com/TheAmirhosssein/class-manager/internal/types"

type School struct {
	ID          types.ID
	Name        string
	OwnerUserID types.ID
}
