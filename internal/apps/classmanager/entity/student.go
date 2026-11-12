package entity

import "github.com/TheAmirhosssein/class-manager/internal/types"

type Student struct {
	ID            types.ID
	ClassID       types.ID
	FirstName     string
	LastName      string
	NationalCode  string
	FatherName    string
	StudentNumber string
}

func (s Student) DisplayName() string {
	if s.FirstName == "" && s.LastName == "" {
		return ""
	}
	if s.LastName == "" || s.LastName == "—" {
		return s.FirstName
	}
	return s.FirstName + " " + s.LastName
}
