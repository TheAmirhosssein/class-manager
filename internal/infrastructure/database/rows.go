package database

import (
	"database/sql"
	"errors"
	"strings"
)

func IsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") || strings.Contains(msg, "constraint failed: unique")
}
