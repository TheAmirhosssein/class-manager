package model

import (
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
)

type StudentJSON struct {
	ID            string `json:"id,omitempty"`
	ClassID       string `json:"classId,omitempty"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Name          string `json:"name"`
	NationalCode  string `json:"nationalCode"`
	FatherName    string `json:"fatherName,omitempty"`
	StudentNumber string `json:"studentNumber,omitempty"`
}

type CreateStudentJSON struct {
	FirstName     string `json:"firstName" binding:"required"`
	LastName      string `json:"lastName" binding:"required"`
	NationalCode  string `json:"nationalCode"`
	FatherName    string `json:"fatherName"`
	StudentNumber string `json:"studentNumber"`
}

type PatchStudentJSON struct {
	ID            string `json:"id,omitempty"`
	FirstName     string `json:"firstName" binding:"required"`
	LastName      string `json:"lastName" binding:"required"`
	NationalCode  string `json:"nationalCode"`
	FatherName    string `json:"fatherName"`
	StudentNumber string `json:"studentNumber"`
}

func NewStudentJSON(s entity.Student) StudentJSON {
	return StudentJSON{
		ID:            strconv.FormatInt(int64(s.ID), 10),
		ClassID:       strconv.FormatInt(int64(s.ClassID), 10),
		FirstName:     s.FirstName,
		LastName:      s.LastName,
		Name:          s.DisplayName(),
		NationalCode:  s.NationalCode,
		FatherName:    s.FatherName,
		StudentNumber: s.StudentNumber,
	}
}
