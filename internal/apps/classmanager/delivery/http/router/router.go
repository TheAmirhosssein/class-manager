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
	classRepo := repository.NewClassRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	disciplineRepo := repository.NewDisciplineRepository(db)
	schoolUC := usecase.NewSchoolUsecase(schoolRepo)
	classUC := usecase.NewClassUsecase(classRepo, schoolRepo)
	studentUC := usecase.NewStudentUsecase(studentRepo, classRepo, schoolRepo)
	sessionUC := usecase.NewSessionUsecase(sessionRepo, classRepo, schoolRepo)
	attendanceUC := usecase.NewAttendanceUsecase(attendanceRepo, sessionRepo, studentRepo, classRepo, schoolRepo)
	disciplineUC := usecase.NewDisciplineUsecase(disciplineRepo, studentRepo, classRepo, schoolRepo)
	schoolH := handler.NewSchoolHandler(schoolUC)
	classH := handler.NewClassHandler(classUC)
	studentH := handler.NewStudentHandler(studentUC)
	sessionH := handler.NewSessionHandler(sessionUC)
	attendanceH := handler.NewAttendanceHandler(attendanceUC)
	disciplineH := handler.NewDisciplineHandler(disciplineUC)

	cookieName := cfg.JWT.CookieName
	api := engine.Group("/api/v1")
	api.Use(middleware.EnsureCSRF())

	secured := api.Group("")
	secured.Use(middleware.PanelAccess(tokens, cookieName))

	secured.GET("/schools", schoolH.List)
	secured.GET("/school", schoolH.Get)
	secured.GET("/classes", classH.List)
	secured.GET("/class", classH.Get)
	secured.GET("/students", studentH.List)
	secured.GET("/student", studentH.Get)
	secured.GET("/sessions", sessionH.List)
	secured.GET("/session", sessionH.Get)
	secured.GET("/attendance", attendanceH.Get)
	secured.GET("/discipline", disciplineH.List)

	writes := secured.Group("")
	writes.Use(middleware.CSRFRequired())
	writes.POST("/school", schoolH.Create)
	writes.PATCH("/school", schoolH.Patch)
	writes.POST("/class", classH.Create)
	writes.PATCH("/class", classH.Patch)
	writes.DELETE("/class", classH.Delete)
	writes.POST("/student", studentH.Create)
	writes.PATCH("/student", studentH.Patch)
	writes.DELETE("/student", studentH.Delete)
	writes.POST("/session", sessionH.Create)
	writes.PATCH("/session", sessionH.Patch)
	writes.DELETE("/session", sessionH.Delete)
	writes.PATCH("/attendance", attendanceH.Patch)
	writes.POST("/discipline", disciplineH.Create)
	writes.PATCH("/discipline", disciplineH.Patch)
	writes.DELETE("/discipline", disciplineH.Delete)
}
