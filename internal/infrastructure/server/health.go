package server

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

func RegisterFrontend(engine *gin.Engine, frontendDir, adminEntryPath string) {
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
	manifest := filepath.Join(frontendDir, "manifest.webmanifest")
	if _, err := os.Stat(manifest); err == nil {
		engine.GET("/manifest.webmanifest", func(c *gin.Context) {
			c.Header("Content-Type", "application/manifest+json; charset=utf-8")
			c.File(manifest)
		})
	}
	swJS := filepath.Join(frontendDir, "sw.js")
	if _, err := os.Stat(swJS); err == nil {
		engine.GET("/sw.js", func(c *gin.Context) {
			c.Header("Content-Type", "application/javascript; charset=utf-8")
			c.Header("Cache-Control", "no-cache")
			c.File(swJS)
		})
	}
	index := filepath.Join(frontendDir, "index.html")
	if adminEntryPath != "" {
		entry := adminEntryPath
		engine.GET("/"+entry, func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/#"+entry)
		})
	}
	engine.GET("/", func(c *gin.Context) {
		c.File(index)
	})
	engine.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Status(http.StatusNotFound)
			return
		}
		p := c.Request.URL.Path
		if p == "/manifest.webmanifest" || p == "/sw.js" || strings.HasPrefix(p, "/assets/") {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(index)
	})
}
