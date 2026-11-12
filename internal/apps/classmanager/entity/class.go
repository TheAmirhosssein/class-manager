package entity

import "github.com/TheAmirhosssein/class-manager/internal/types"

type Class struct {
	ID           types.ID
	SchoolID     types.ID
	Title        string
	Grade        string
	Room         string
	StudentCount int
}
