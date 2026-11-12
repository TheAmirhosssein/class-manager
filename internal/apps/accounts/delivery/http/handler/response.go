package handler

import (
	sharedmessages "github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/gin-gonic/gin"
)

func writeMessage(c *gin.Context, msg sharedmessages.Message) {
	c.JSON(msg.HttpCode(), gin.H{"message": msg.Description})
}
