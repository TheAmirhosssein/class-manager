package validation

import "strings"

var validAttendanceStatuses = map[string]struct{}{
	"present": {},
	"absent":  {},
	"late":    {},
	"excused": {},
}

func NormalizeAttendanceStatus(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func ValidAttendanceStatus(s string) bool {
	_, ok := validAttendanceStatuses[NormalizeAttendanceStatus(s)]
	return ok
}
