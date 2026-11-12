package validation

import (
	"strings"
	"unicode/utf8"
)

const (
	minPasswordLen = 8
	minNameLen     = 2
	maxNameLen     = 120
)

func ValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at < 1 || at == len(email)-1 {
		return false
	}
	domain := email[at+1:]
	return strings.Contains(domain, ".")
}

func ValidUserName(name string) bool {
	name = strings.TrimSpace(name)
	n := utf8.RuneCountInString(name)
	return n >= minNameLen && n <= maxNameLen
}

func ValidPassword(password string) bool {
	return utf8.RuneCountInString(password) >= minPasswordLen
}
