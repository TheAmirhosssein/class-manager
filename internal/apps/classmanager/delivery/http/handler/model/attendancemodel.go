package model

import (
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
)

type AttendanceRecordJSON struct {
	StudentID string `json:"studentId"`
	Status    string `json:"status"`
}

type AttendanceJSON struct {
	SessionID string                 `json:"sessionId"`
	ClassID   string                 `json:"classId"`
	Records   []AttendanceRecordJSON `json:"records"`
}

type SetAttendanceJSON struct {
	Status string `json:"status" binding:"required"`
}

func NewAttendanceJSON(roll entity.AttendanceRoll) AttendanceJSON {
	out := AttendanceJSON{
		SessionID: strconv.FormatInt(int64(roll.SessionID), 10),
		ClassID:   strconv.FormatInt(int64(roll.ClassID), 10),
		Records:   make([]AttendanceRecordJSON, 0, len(roll.Records)),
	}
	for _, r := range roll.Records {
		out.Records = append(out.Records, AttendanceRecordJSON{
			StudentID: strconv.FormatInt(int64(r.StudentID), 10),
			Status:    r.Status,
		})
	}
	return out
}

func NewAttendanceRecordJSON(r entity.AttendanceRecord) AttendanceRecordJSON {
	return AttendanceRecordJSON{
		StudentID: strconv.FormatInt(int64(r.StudentID), 10),
		Status:    r.Status,
	}
}
