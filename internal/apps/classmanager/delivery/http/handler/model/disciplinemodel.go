package model

import (
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
)

type DisciplineJSON struct {
	ID          string `json:"id,omitempty"`
	ClassID     string `json:"classId,omitempty"`
	StudentID   string `json:"studentId,omitempty"`
	StudentName string `json:"studentName,omitempty"`
	Type        string `json:"type"`
	Note        string `json:"note"`
	RecordDate  string `json:"recordDate"`
}

type CreateDisciplineJSON struct {
	StudentID  string `json:"studentId"`
	Type       string `json:"type"`
	RecordDate string `json:"recordDate"`
	Note       string `json:"note"`
}

type PatchDisciplineJSON struct {
	ID         string `json:"id,omitempty"`
	ClassID    string `json:"classId,omitempty"`
	StudentID  string `json:"studentId,omitempty"`
	Type       string `json:"type"`
	RecordDate string `json:"recordDate"`
	Note       string `json:"note"`
}

func NewDisciplineJSON(d entity.DisciplineRecord) DisciplineJSON {
	return DisciplineJSON{
		ID:          strconv.FormatInt(int64(d.ID), 10),
		ClassID:     strconv.FormatInt(int64(d.ClassID), 10),
		StudentID:   strconv.FormatInt(int64(d.StudentID), 10),
		StudentName: d.StudentName,
		Type:        d.Type,
		Note:        d.Note,
		RecordDate:  d.RecordDate,
	}
}
