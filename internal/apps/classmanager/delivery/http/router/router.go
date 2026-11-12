package router

import (
	"database/sql"

	"github.com/TheAmirhosssein/class-manager/config"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/delivery/http/handler"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/usecase"
	"github.com/TheAmirhosssein/class-manager/internal/httpconfig/middleware"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine, cfg *config.Config, db *sql.DB, tokens *jwt.Service) {
	schoolRepo := repository.NewSchoolRepository(db)
	schoolUC := usecase.NewSchoolUsecase(schoolRepo)
	schoolH := handler.NewSchoolHandler(schoolUC)

	cookieName := cfg.JWT.CookieName
	api := engine.Group("/api/v1")
	api.Use(middleware.EnsureCSRF())

	secured := api.Group("")
	secured.Use(middleware.PanelAccess(tokens, cookieName))

	secured.GET("/schools", schoolH.List)
	secured.GET("/school", schoolH.Get)

	writes := secured.Group("")
	writes.Use(middleware.CSRFRequired())
	writes.POST("/school", schoolH.Create)
	writes.PATCH("/school", schoolH.Patch)
}
