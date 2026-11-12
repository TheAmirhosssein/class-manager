package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      App      `mapstructure:"app"`
	HTTP     HTTP     `mapstructure:"http"`
	Database Database `mapstructure:"database"`
	Frontend Frontend `mapstructure:"frontend"`
}

type App struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type HTTP struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Database struct {
	Driver string `mapstructure:"driver"`
	Path   string `mapstructure:"path"`
}

type Frontend struct {
	Dir string `mapstructure:"dir"`
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
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
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
