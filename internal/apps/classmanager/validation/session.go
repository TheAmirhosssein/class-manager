package validation

import (
	"regexp"
	"strings"
	"time"
)

var clockRE = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func tehranLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		return time.FixedZone("Asia/Tehran", 3*3600+30*60)
	}
	return loc
}

func NormalizeSessionDate(s string) string {
	return strings.TrimSpace(s)
}

func ValidSessionDate(date string) bool {
	date = NormalizeSessionDate(date)
	if date == "" {
		return false
	}
	_, err := time.ParseInLocation("2006-01-02", date, tehranLoc())
	return err == nil
}

func ValidSessionDay(day int) bool {
	return day >= 0 && day <= 6
}

func WeekdayTehran(date string) (int, bool) {
	date = NormalizeSessionDate(date)
	t, err := time.ParseInLocation("2006-01-02", date, tehranLoc())
	if err != nil {
		return 0, false
	}
	return (int(t.Weekday()) + 1) % 7, true
}

func DayMatchesSessionDate(day int, date string) bool {
	want, ok := WeekdayTehran(date)
	if !ok {
		return false
	}
	return day == want
}

func NormalizeClock(s string) string {
	return strings.TrimSpace(s)
}

func ValidClock(s string) bool {
	return clockRE.MatchString(NormalizeClock(s))
}

func ValidSessionTimeRange(start, end string) bool {
	start = NormalizeClock(start)
	end = NormalizeClock(end)
	if !ValidClock(start) || !ValidClock(end) {
		return false
	}
	return start < end
}

func NormalizeSessionSubject(s string) string {
	return strings.TrimSpace(s)
}

func ValidSessionSubject(s string) bool {
	s = NormalizeSessionSubject(s)
	return len(s) <= 100
}
