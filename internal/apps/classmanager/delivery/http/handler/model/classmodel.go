package model

import (
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
)

type ClassJSON struct {
	ID           string `json:"id,omitempty"`
	SchoolID     string `json:"schoolId,omitempty"`
	Title        string `json:"title"`
	Grade        string `json:"grade"`
	Room         string `json:"room,omitempty"`
	StudentCount int    `json:"studentCount,omitempty"`
}

type CreateClassJSON struct {
	Title string `json:"title" binding:"required"`
	Grade string `json:"grade" binding:"required"`
	Room  string `json:"room"`
}

type PatchClassJSON struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title" binding:"required"`
	Grade string `json:"grade" binding:"required"`
	Room  string `json:"room"`
}

func NewClassJSON(c entity.Class) ClassJSON {
	return ClassJSON{
		ID:           strconv.FormatInt(int64(c.ID), 10),
		SchoolID:     strconv.FormatInt(int64(c.SchoolID), 10),
		Title:        c.Title,
		Grade:        c.Grade,
		Room:         c.Room,
		StudentCount: c.StudentCount,
	}
}
