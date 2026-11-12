package validation

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	minPasswordLen = 8
	minNameLen     = 2
	maxNameLen     = 120
)

func NormalizeMobile(raw string) string {
	raw = strings.TrimSpace(raw)
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if strings.HasPrefix(s, "98") && len(s) == 12 {
		return "0" + s[2:]
	}
	if strings.HasPrefix(s, "9") && len(s) == 10 {
		return "0" + s
	}
	return s
}

func ValidMobile(mobile string) bool {
	m := NormalizeMobile(mobile)
	return len(m) == 11 && strings.HasPrefix(m, "09")
}

func ValidUserName(name string) bool {
	name = strings.TrimSpace(name)
	n := utf8.RuneCountInString(name)
	return n >= minNameLen && n <= maxNameLen
}

func ValidPassword(password string) bool {
	return utf8.RuneCountInString(password) >= minPasswordLen
}
