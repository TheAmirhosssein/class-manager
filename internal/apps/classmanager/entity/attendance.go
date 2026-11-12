package entity

import "github.com/TheAmirhosssein/class-manager/internal/types"

type AttendanceRecord struct {
	StudentID types.ID
	Status    string
}

type AttendanceRoll struct {
	SessionID types.ID
	ClassID   types.ID
	Records   []AttendanceRecord
}
