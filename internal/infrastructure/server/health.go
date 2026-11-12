package server

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func RegisterHealth(engine *gin.Engine, db *sql.DB) {
	engine.GET("/healthz", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "db": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}

func RegisterFrontend(engine *gin.Engine, frontendDir string) {
	if frontendDir == "" {
		return
	}
	if _, err := os.Stat(frontendDir); err != nil {
		return
	}
	assetsDir := filepath.Join(frontendDir, "assets")
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		engine.Static("/assets", assetsDir)
	}
	index := filepath.Join(frontendDir, "index.html")
	engine.GET("/", func(c *gin.Context) {
		c.File(index)
	})
	engine.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(index)
	})
}
