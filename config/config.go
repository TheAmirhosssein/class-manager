package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App      App      `yaml:"app"`
	HTTP     HTTP     `yaml:"http"`
	Database Database `yaml:"database"`
	Frontend Frontend `yaml:"frontend"`
	JWT      JWT      `yaml:"jwt"`
}

type JWT struct {
	Secret     string        `yaml:"secret" env:"JWT_SECRET"`
	AccessTTL  time.Duration `yaml:"access_ttl" env:"JWT_ACCESS_TTL" env-default:"24h"`
	CookieName string        `yaml:"cookie_name" env:"JWT_COOKIE_NAME" env-default:"access_token"`
}

type App struct {
	Name    string `yaml:"name" env:"APP_NAME" env-default:"classino"`
	Version string `yaml:"version" env:"APP_VERSION" env-default:"0.1.0"`
}

type HTTP struct {
	Host string `yaml:"host" env:"HTTP_HOST" env-default:"127.0.0.1"`
	Port int    `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
}

type Database struct {
	Driver string `yaml:"driver" env:"DATABASE_DRIVER" env-default:"sqlite"`
	Path   string `yaml:"path" env:"DATABASE_PATH" env-default:"data/classino.db"`
}

type Frontend struct {
	Dir string `yaml:"dir" env:"FRONTEND_DIR" env-default:"frontend"`
}

var (
	loaded   sync.Once
	instance *Config
)

func GetConfig() *Config {
	loaded.Do(func() {
		if err := load(); err != nil {
			panic(err)
		}
	})
	return instance
}

func load() error {
	var c Config

	configPath, err := resolveConfigPath()
	if err != nil {
		return err
	}
	if err := cleanenv.ReadConfig(configPath, &c); err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	if envPath, err := resolveEnvPath(); err == nil {
		if err := cleanenv.ReadConfig(envPath, &c); err != nil {
			return fmt.Errorf("read env file: %w", err)
		}
	}

	if c.JWT.AccessTTL <= 0 {
		c.JWT.AccessTTL = 24 * time.Hour
	}

	if c.Database.Path != "" && !filepath.IsAbs(c.Database.Path) {
		if root, err := projectRoot(); err == nil {
			c.Database.Path = filepath.Join(root, c.Database.Path)
		}
	}
	if c.Frontend.Dir != "" && !filepath.IsAbs(c.Frontend.Dir) {
		if root, err := projectRoot(); err == nil {
			c.Frontend.Dir = filepath.Join(root, c.Frontend.Dir)
		}
	}

	instance = &c
	return nil
}

func resolveConfigPath() (string, error) {
	candidates := []string{
		"config/config.yaml",
		"config.yaml",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("config file not found (tried config/config.yaml and config.yaml)")
}

func resolveEnvPath() (string, error) {
	if _, err := os.Stat(".env"); err != nil {
		return "", err
	}
	return ".env", nil
}

func projectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return wd, nil
}

func InTestMode() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") || strings.Contains(arg, "/test.") {
			return true
		}
	}
	return false
}
