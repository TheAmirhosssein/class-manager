package handler

import (
	"net/http"
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/delivery/http/handler/model"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/usecase"
	"github.com/TheAmirhosssein/class-manager/internal/httpconfig/middleware"
	sharedmessages "github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth       *usecase.AuthUsecase
	cookieName string
	tokenTTL   time.Duration
}

func NewAuthHandler(auth *usecase.AuthUsecase, cookieName string, tokenTTL time.Duration) *AuthHandler {
	if cookieName == "" {
		cookieName = middleware.DefaultCookieName
	}
	return &AuthHandler{auth: auth, cookieName: cookieName, tokenTTL: tokenTTL}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var body model.LoginModel
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(sharedmessages.BadRequest.HttpCode(), gin.H{"message": sharedmessages.BadRequest.Description})
		return
	}
	result, msg := h.auth.LoginTeacher(c.Request.Context(), body.Email, body.Password)
	if msg.Error != nil {
		c.JSON(msg.HttpCode(), gin.H{"message": msg.Description})
		return
	}
	h.setTokenCookie(c, result.Token)
	c.JSON(http.StatusOK, model.LoginResponse{User: result.User, Token: result.Token})
}

func (h *AuthHandler) Session(c *gin.Context) {
	raw := middleware.BearerToken(c, h.cookieName)
	if raw == "" {
		c.JSON(message.SessionInvalid.HttpCode(), gin.H{"message": message.SessionInvalid.Description})
		return
	}
	user, msg := h.auth.Session(c.Request.Context(), raw)
	if msg.Error != nil {
		c.JSON(msg.HttpCode(), gin.H{"message": msg.Description})
		return
	}
	c.JSON(http.StatusOK, model.SessionResponse{User: user})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearTokenCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) setTokenCookie(c *gin.Context, token string) {
	maxAge := int(h.tokenTTL.Seconds())
	if maxAge <= 0 {
		maxAge = 86400
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cookieName, token, maxAge, "/", "", false, true)
}

func (h *AuthHandler) clearTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cookieName, "", -1, "/", "", false, true)
}
