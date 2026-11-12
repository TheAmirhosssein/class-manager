package handler

import (
	"net/http"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/delivery/http/handler/model"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/usecase"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	sharedmessages "github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/gin-gonic/gin"
)

type DisciplineHandler struct {
	discipline *usecase.DisciplineUsecase
}

func NewDisciplineHandler(discipline *usecase.DisciplineUsecase) *DisciplineHandler {
	return &DisciplineHandler{discipline: discipline}
}

func (h *DisciplineHandler) List(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	list, msg := h.discipline.List(c.Request.Context(), params.ListDisciplineParams{
		Actor:     actor,
		SchoolID:  schoolIDQuery(c),
		ClassID:   classIDQuery(c),
		StudentID: studentIDQuery(c),
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	out := make([]model.DisciplineJSON, 0, len(list))
	for _, d := range list {
		out = append(out, model.NewDisciplineJSON(d))
	}
	c.JSON(http.StatusOK, out)
}

func (h *DisciplineHandler) Create(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	classID := classIDQuery(c)
	if !classID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.CreateDisciplineJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	record, msg := h.discipline.Create(c.Request.Context(), params.CreateDisciplineParams{
		Actor:      actor,
		ClassID:    classID,
		StudentID:  parseID(body.StudentID),
		Type:       body.Type,
		RecordDate: body.RecordDate,
		Note:       body.Note,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewDisciplineJSON(record))
}

func (h *DisciplineHandler) Patch(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.PatchDisciplineJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	recordID := disciplineIDQuery(c)
	if !recordID.Valid() && body.ID != "" {
		recordID = parseID(body.ID)
	}
	if !recordID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	classID := types.ID(0)
	if body.ClassID != "" {
		classID = parseID(body.ClassID)
	}
	studentID := types.ID(0)
	if body.StudentID != "" {
		studentID = parseID(body.StudentID)
	}
	record, msg := h.discipline.Update(c.Request.Context(), params.UpdateDisciplineParams{
		Actor:      actor,
		RecordID:   recordID,
		ClassID:    classID,
		StudentID:  studentID,
		Type:       body.Type,
		RecordDate: body.RecordDate,
		Note:       body.Note,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewDisciplineJSON(record))
}

func (h *DisciplineHandler) Delete(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	recordID := disciplineIDQuery(c)
	if !recordID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	msg := h.discipline.Delete(c.Request.Context(), params.DeleteDisciplineParams{
		Actor:    actor,
		RecordID: recordID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), gin.H{"message": msg.Description})
}

func disciplineIDQuery(c *gin.Context) types.ID {
	return parseID(c.Query("disciplineId"))
}
