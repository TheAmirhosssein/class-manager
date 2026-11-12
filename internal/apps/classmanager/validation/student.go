package validation

import (
	"strings"
	"unicode"
)

func NormalizeStudentFirstName(s string) string {
	return strings.TrimSpace(s)
}

func NormalizeStudentLastName(s string) string {
	return strings.TrimSpace(s)
}

func NormalizeStudentFatherName(s string) string {
	return strings.TrimSpace(s)
}

func NormalizeStudentNumber(s string) string {
	return strings.TrimSpace(s)
}

func NormalizeNationalCode(raw string) string {
	s := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r >= '۰' && r <= '۹' {
			return '0' + (r - '۰')
		}
		return -1
	}, raw)
	return s
}

func ValidStudentFirstName(s string) bool {
	s = NormalizeStudentFirstName(s)
	if len(s) < 2 || len(s) > 60 {
		return false
	}
	return hasLetter(s)
}

func ValidStudentLastName(s string) bool {
	s = NormalizeStudentLastName(s)
	if len(s) < 2 || len(s) > 60 {
		return false
	}
	return hasLetter(s)
}

func ValidStudentFatherName(s string) bool {
	s = NormalizeStudentFatherName(s)
	return len(s) <= 60
}

func ValidStudentNumber(s string) bool {
	s = NormalizeStudentNumber(s)
	return len(s) <= 20
}

func ValidNationalCode(code string) bool {
	code = NormalizeNationalCode(code)
	if code == "" {
		return true
	}
	if len(code) != 10 {
		return false
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}
	if code == strings.Repeat(code[:1], 10) {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		d := int(code[i] - '0')
		sum += d * (10 - i)
	}
	rem := sum % 11
	check := int(code[9] - '0')
	if rem < 2 {
		return check == rem
	}
	return check == 11-rem
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
