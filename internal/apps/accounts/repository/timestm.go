package repository

import (
	"time"
)

const sqliteTimeLayout = "2006-01-02T15:04:05.999999Z"

func parseSQLiteTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(sqliteTimeLayout, s)
}

func formatSQLiteTime(t time.Time) string {
	return t.UTC().Format(sqliteTimeLayout)
}
