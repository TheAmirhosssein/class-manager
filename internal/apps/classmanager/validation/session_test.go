package validation

import "testing"

func TestWeekdayTehranSaturday(t *testing.T) {
	day, ok := WeekdayTehran("2025-11-15")
	if !ok || day != 0 {
		t.Fatalf("2025-11-15 (Saturday) want day 0, got %d ok=%v", day, ok)
	}
}

func TestDayMatchesSessionDate(t *testing.T) {
	if !DayMatchesSessionDate(0, "2025-11-15") {
		t.Fatal("Saturday should match")
	}
	if DayMatchesSessionDate(6, "2025-11-15") {
		t.Fatal("Friday should not match Saturday")
	}
}
