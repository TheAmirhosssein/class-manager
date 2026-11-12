package model

import (
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
)

type SessionJSON struct {
	ID          string `json:"id,omitempty"`
	ClassID     string `json:"classId,omitempty"`
	SchoolID    string `json:"schoolId,omitempty"`
	SessionDate string `json:"sessionDate"`
	Day         int    `json:"day"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Subject     string `json:"subject,omitempty"`
}

type CreateSessionJSON struct {
	SessionDate string `json:"sessionDate"`
	Day         int    `json:"day"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Subject     string `json:"subject"`
}

type PatchSessionJSON struct {
	ID          string `json:"id,omitempty"`
	SessionDate string `json:"sessionDate"`
	Day         int    `json:"day"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Subject     string `json:"subject"`
}

func NewSessionJSON(s entity.Session) SessionJSON {
	return SessionJSON{
		ID:          strconv.FormatInt(int64(s.ID), 10),
		ClassID:     strconv.FormatInt(int64(s.ClassID), 10),
		SchoolID:    strconv.FormatInt(int64(s.SchoolID), 10),
		SessionDate: s.SessionDate,
		Day:         s.Day,
		StartTime:   s.StartTime,
		EndTime:     s.EndTime,
		Subject:     s.Subject,
	}
}
