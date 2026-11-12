package handler

import (
	"net/http"
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/delivery/http/handler/model"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/usecase"
	"github.com/TheAmirhosssein/class-manager/internal/httpconfig/middleware"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	sharedmessages "github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/gin-gonic/gin"
)

type SchoolHandler struct {
	schools *usecase.SchoolUsecase
}

func NewSchoolHandler(schools *usecase.SchoolUsecase) *SchoolHandler {
	return &SchoolHandler{schools: schools}
}

func (h *SchoolHandler) List(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	list, msg := h.schools.List(c.Request.Context(), params.ListSchoolsParams{Actor: actor})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	out := make([]model.SchoolJSON, 0, len(list))
	for _, s := range list {
		out = append(out, model.NewSchoolJSON(s))
	}
	c.JSON(http.StatusOK, out)
}

func (h *SchoolHandler) Get(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	school, msg := h.schools.Read(c.Request.Context(), params.ReadSchoolParams{
		Actor:    actor,
		SchoolID: schoolIDQuery(c),
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewSchoolJSON(school))
}

func (h *SchoolHandler) Create(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.SchoolJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	school, msg := h.schools.Create(c.Request.Context(), params.CreateSchoolParams{
		Actor: actor,
		Name:  body.Name,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewSchoolJSON(school))
}

func (h *SchoolHandler) Patch(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.PatchSchoolJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	schoolID := schoolIDQuery(c)
	if !schoolID.Valid() && body.ID != "" {
		schoolID = parseID(body.ID)
	}
	school, msg := h.schools.UpdateName(c.Request.Context(), params.UpdateSchoolParams{
		Actor:    actor,
		SchoolID: schoolID,
		Name:     body.Name,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewSchoolJSON(school))
}

func ownerActor(c *gin.Context) (params.OwnerActor, bool) {
	id, ok := middleware.TeacherUserID(c)
	if !ok {
		return params.OwnerActor{}, false
	}
	return params.OwnerActor{UserID: types.ID(id)}, true
}

func schoolIDQuery(c *gin.Context) types.ID {
	return parseID(c.Query("schoolId"))
}

func parseID(s string) types.ID {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return types.ID(n)
}

func writeMessage(c *gin.Context, msg sharedmessages.Message) {
	c.JSON(msg.HttpCode(), gin.H{"message": msg.Description})
}
