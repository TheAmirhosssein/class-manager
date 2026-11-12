package handler

import (
	"net/http"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/delivery/http/handler/model"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/usecase"
	sharedmessages "github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct {
	attendance *usecase.AttendanceUsecase
}

func NewAttendanceHandler(attendance *usecase.AttendanceUsecase) *AttendanceHandler {
	return &AttendanceHandler{attendance: attendance}
}

func (h *AttendanceHandler) Get(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	sessionID := sessionIDQuery(c)
	if !sessionID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	roll, msg := h.attendance.Read(c.Request.Context(), params.ReadAttendanceParams{
		Actor:           actor,
		WeeklySessionID: sessionID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(http.StatusOK, model.NewAttendanceJSON(roll))
}

func (h *AttendanceHandler) Patch(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	sessionID := sessionIDQuery(c)
	studentID := studentIDQuery(c)
	if !sessionID.Valid() || !studentID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.SetAttendanceJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	rec, msg := h.attendance.SetRecord(c.Request.Context(), params.SetAttendanceRecordParams{
		Actor:           actor,
		WeeklySessionID: sessionID,
		StudentID:       studentID,
		Status:          body.Status,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewAttendanceRecordJSON(rec))
}
