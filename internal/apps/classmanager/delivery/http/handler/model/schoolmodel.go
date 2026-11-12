package model

import (
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
)

type SchoolJSON struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type PatchSchoolJSON struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name" binding:"required"`
}

func NewSchoolJSON(s entity.School) SchoolJSON {
	return SchoolJSON{
		ID:   strconv.FormatInt(int64(s.ID), 10),
		Name: s.Name,
	}
}
