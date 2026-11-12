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

type SessionHandler struct {
	sessions *usecase.SessionUsecase
}

func NewSessionHandler(sessions *usecase.SessionUsecase) *SessionHandler {
	return &SessionHandler{sessions: sessions}
}

func (h *SessionHandler) List(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	list, msg := h.sessions.List(c.Request.Context(), params.ListSessionsParams{
		Actor:    actor,
		SchoolID: schoolIDQuery(c),
		ClassID:  classIDQuery(c),
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	out := make([]model.SessionJSON, 0, len(list))
	for _, s := range list {
		out = append(out, model.NewSessionJSON(s))
	}
	c.JSON(http.StatusOK, out)
}

func (h *SessionHandler) Get(c *gin.Context) {
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
	session, msg := h.sessions.Read(c.Request.Context(), params.ReadSessionParams{
		Actor:     actor,
		SessionID: sessionID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewSessionJSON(session))
}

func (h *SessionHandler) Create(c *gin.Context) {
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
	var body model.CreateSessionJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	session, msg := h.sessions.Create(c.Request.Context(), params.CreateSessionParams{
		Actor:       actor,
		SchoolID:    schoolIDQuery(c),
		ClassID:     classID,
		SessionDate: body.SessionDate,
		Day:         body.Day,
		Start:       body.StartTime,
		End:         body.EndTime,
		Subject:     body.Subject,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewSessionJSON(session))
}

func (h *SessionHandler) Patch(c *gin.Context) {
	actor, ok := ownerActor(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.PatchSessionJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	sessionID := sessionIDQuery(c)
	if !sessionID.Valid() && body.ID != "" {
		sessionID = parseID(body.ID)
	}
	if !sessionID.Valid() {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	session, msg := h.sessions.Update(c.Request.Context(), params.UpdateSessionParams{
		Actor:       actor,
		SessionID:   sessionID,
		SessionDate: body.SessionDate,
		Day:         body.Day,
		Start:       body.StartTime,
		End:         body.EndTime,
		Subject:     body.Subject,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewSessionJSON(session))
}

func (h *SessionHandler) Delete(c *gin.Context) {
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
	msg := h.sessions.Delete(c.Request.Context(), params.DeleteSessionParams{
		Actor:     actor,
		SessionID: sessionID,
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), gin.H{"message": msg.Description})
}

func sessionIDQuery(c *gin.Context) types.ID {
	return parseID(c.Query("sessionId"))
}
