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

type StudentHandler struct {
	students *usecase.StudentUsecase
}

func NewStudentHandler(students *usecase.StudentUsecase) *StudentHandler {
	return &StudentHandler{students: students}
}

func (h *StudentHandler) List(c *gin.Context) {
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
	list, msg := h.students.List(c.Request.Context(), params.ListStudentsParams{
		Actor:   actor,
		ClassID: classID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	out := make([]model.StudentJSON, 0, len(list))
	for _, s := range list {
		out = append(out, model.NewStudentJSON(s))
	}
	c.JSON(http.StatusOK, out)
}

func (h *StudentHandler) Get(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	studentID := studentIDQuery(c)
	if !studentID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	student, msg := h.students.Read(c.Request.Context(), params.ReadStudentParams{
		Actor:     actor,
		StudentID: studentID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewStudentJSON(student))
}

func (h *StudentHandler) Create(c *gin.Context) {
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
	var body model.CreateStudentJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	student, msg := h.students.Create(c.Request.Context(), params.CreateStudentParams{
		Actor:         actor,
		ClassID:       classID,
		FirstName:     body.FirstName,
		LastName:      body.LastName,
		NationalCode:  body.NationalCode,
		FatherName:    body.FatherName,
		StudentNumber: body.StudentNumber,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewStudentJSON(student))
}

func (h *StudentHandler) Patch(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.PatchStudentJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	studentID := studentIDQuery(c)
	if !studentID.Valid() && body.ID != "" {
		studentID = parseID(body.ID)
	}
	if !studentID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	student, msg := h.students.Update(c.Request.Context(), params.UpdateStudentParams{
		Actor:         actor,
		StudentID:     studentID,
		FirstName:     body.FirstName,
		LastName:      body.LastName,
		NationalCode:  body.NationalCode,
		FatherName:    body.FatherName,
		StudentNumber: body.StudentNumber,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewStudentJSON(student))
}

func (h *StudentHandler) Delete(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	studentID := studentIDQuery(c)
	if !studentID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	msg := h.students.Delete(c.Request.Context(), params.DeleteStudentParams{
		Actor:     actor,
		StudentID: studentID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), gin.H{"message": msg.Description})
}

func studentIDQuery(c *gin.Context) types.ID {
	return parseID(c.Query("studentId"))
}
