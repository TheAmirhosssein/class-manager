package router

import (
	"regexp"
	"strings"

	"github.com/TheAmirhosssein/class-manager/config"
)

var adminPathPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}[a-z0-9]$|^[a-z0-9]$`)

func normalizeAdminPathSecret(raw string) string {
	s := strings.Trim(strings.TrimSpace(raw), "/")
	if s == "" || !adminPathPattern.MatchString(s) {
		return "sys-ingress"
	}
	return s
}

// EntryPath returns the configured secret path segment for the admin console.
func EntryPath(cfg *config.Config) string {
	if cfg == nil {
		return "sys-ingress"
	}
	return normalizeAdminPathSecret(cfg.Admin.PathSecret)
}
