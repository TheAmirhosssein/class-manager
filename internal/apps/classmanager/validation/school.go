package validation

import "strings"

func ValidSchoolName(name string) bool {
	name = strings.TrimSpace(name)
	return len(name) >= 2 && len(name) <= 80
}

func NormalizeSchoolName(name string) string {
	return strings.TrimSpace(name)
}
