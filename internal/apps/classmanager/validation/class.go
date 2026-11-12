package validation

import "strings"

var validGrades = map[string]struct{}{
	"10": {},
	"11": {},
	"12": {},
}

func ValidClassTitle(title string) bool {
	title = strings.TrimSpace(title)
	return len(title) >= 2 && len(title) <= 100
}

func NormalizeClassTitle(title string) string {
	return strings.TrimSpace(title)
}

func ValidClassGrade(grade string) bool {
	grade = strings.TrimSpace(grade)
	_, ok := validGrades[grade]
	return ok
}

func NormalizeClassGrade(grade string) string {
	return strings.TrimSpace(grade)
}

func ValidClassRoom(room string) bool {
	room = strings.TrimSpace(room)
	return len(room) <= 80
}

func NormalizeClassRoom(room string) string {
	return strings.TrimSpace(room)
}
