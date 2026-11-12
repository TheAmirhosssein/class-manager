package handler

import (
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/delivery/http/handler/model"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/usecase"
	"github.com/TheAmirhosssein/class-manager/internal/httpconfig/middleware"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	sharedmessages "github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users *usecase.UserUsecase
}

func NewUserHandler(users *usecase.UserUsecase) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Create(c *gin.Context) {
	adminID, ok := middleware.AdminActorID(c)
	if !ok {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	var body model.CreateUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeMessage(c, sharedmessages.BadRequest)
		return
	}
	created, msg := h.users.CreateByAdmin(c.Request.Context(), params.CreateUserUsecaseParams{
		ActorAdminID: types.ID(adminID),
		CreateUserParams: params.CreateUserParams{
			Mobile:   body.Mobile,
			Password: body.Password,
			Name:     body.Name,
		},
	})
	if msg.Error != nil {
		writeMessage(c, msg)
		return
	}
	c.JSON(msg.HttpCode(), model.NewCreatedUserJSON(created))
}
