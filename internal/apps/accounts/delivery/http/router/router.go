package router

import (
	"database/sql"

	"github.com/TheAmirhosssein/class-manager/config"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/delivery/http/handler"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/usecase"
	"github.com/TheAmirhosssein/class-manager/internal/httpconfig/middleware"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/hash"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine, cfg *config.Config, db *sql.DB, tokens *jwt.Service) {
	h := hash.New()
	userRepo := repository.NewUserRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	users := usecase.NewUserUsecase(userRepo, adminRepo, h)
	authUC := usecase.NewAuthUsecase(users, userRepo, h, tokens)
	authH := handler.NewAuthHandler(authUC, cfg.JWT.CookieName, cfg.JWT.AccessTTL)

	api := engine.Group("/api/v1")
	api.Use(middleware.EnsureCSRF())
	auth := api.Group("/auth")
	auth.POST("/login", authH.Login)
	auth.GET("/session", authH.Session)
	auth.POST("/logout", authH.Logout)
}
