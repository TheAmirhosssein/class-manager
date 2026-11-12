package entity

import "github.com/TheAmirhosssein/class-manager/internal/types"

type DisciplineRecord struct {
	ID          types.ID
	ClassID     types.ID
	StudentID   types.ID
	StudentName string
	Type        string
	Note        string
	RecordDate  string
}
