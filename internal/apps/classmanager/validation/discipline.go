package validation

import "strings"

func NormalizeDisciplineNote(note string) string {
	return strings.TrimSpace(note)
}

func ValidDisciplineNote(note string) bool {
	n := NormalizeDisciplineNote(note)
	return len(n) >= 3 && len(n) <= 1000
}

func NormalizeDisciplineType(t string) string {
	return strings.TrimSpace(t)
}

func ValidDisciplineType(t string) bool {
	switch NormalizeDisciplineType(t) {
	case "positive", "attention":
		return true
	default:
		return false
	}
}

func NormalizeDisciplineDate(d string) string {
	return strings.TrimSpace(d)
}

func ValidDisciplineDate(d string) bool {
	return ValidSessionDate(NormalizeDisciplineDate(d))
}
