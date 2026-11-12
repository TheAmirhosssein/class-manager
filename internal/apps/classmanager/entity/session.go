package entity

import "github.com/TheAmirhosssein/class-manager/internal/types"

type Session struct {
	ID          types.ID
	ClassID     types.ID
	SchoolID    types.ID
	SessionDate string
	Day         int
	StartTime   string
	EndTime     string
	Subject     string
}
