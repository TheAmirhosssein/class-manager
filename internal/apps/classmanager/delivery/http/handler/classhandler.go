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

type ClassHandler struct {
	classes *usecase.ClassUsecase
}

func NewClassHandler(classes *usecase.ClassUsecase) *ClassHandler {
	return &ClassHandler{classes: classes}
}

func (h *ClassHandler) List(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	list, msg := h.classes.List(c.Request.Context(), params.ListClassesParams{
		Actor:    actor,
		SchoolID: schoolIDQuery(c),
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	out := make([]model.ClassJSON, 0, len(list))
	for _, cl := range list {
		out = append(out, model.NewClassJSON(cl))
	}
	c.JSON(http.StatusOK, out)
}

func (h *ClassHandler) Get(c *gin.Context) {
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
	class, msg := h.classes.Read(c.Request.Context(), params.ReadClassParams{
		Actor:   actor,
		ClassID: classID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewClassJSON(class))
}

func (h *ClassHandler) Create(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.CreateClassJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	class, msg := h.classes.Create(c.Request.Context(), params.CreateClassParams{
		Actor:    actor,
		SchoolID: schoolIDQuery(c),
		Title:    body.Title,
		Grade:    body.Grade,
		Room:     body.Room,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewClassJSON(class))
}

func (h *ClassHandler) Patch(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.PatchClassJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	classID := classIDQuery(c)
	if !classID.Valid() && body.ID != "" {
		classID = parseID(body.ID)
	}
	if !classID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	class, msg := h.classes.Update(c.Request.Context(), params.UpdateClassParams{
		Actor:    actor,
		ClassID:  classID,
		SchoolID: schoolIDQuery(c),
		Title:    body.Title,
		Grade:    body.Grade,
		Room:     body.Room,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewClassJSON(class))
}

func (h *ClassHandler) Delete(c *gin.Context) {
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
	msg := h.classes.Delete(c.Request.Context(), params.DeleteClassParams{
		Actor:   actor,
		ClassID: classID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), gin.H{"message": msg.Description})
}

func classIDQuery(c *gin.Context) types.ID {
	return parseID(c.Query("classId"))
}
