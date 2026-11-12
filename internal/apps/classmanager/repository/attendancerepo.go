package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/database"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type AttendanceRepository interface {
	EnsureForWeeklySession(ctx context.Context, weeklySessionID, classID types.ID, sessionDate string) (types.ID, error)
	ListRecords(ctx context.Context, weeklySessionID types.ID) ([]entity.AttendanceRecord, error)
	UpsertRecord(ctx context.Context, weeklySessionID types.ID, studentID types.ID, status string) error
}

type attendanceRepo struct {
	db *sql.DB
}

var _ AttendanceRepository = (*attendanceRepo)(nil)

func NewAttendanceRepository(db *sql.DB) AttendanceRepository {
	return &attendanceRepo{db: db}
}

func (r *attendanceRepo) EnsureForWeeklySession(ctx context.Context, weeklySessionID, classID types.ID, sessionDate string) (types.ID, error) {
	var id types.ID
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM attendance_sessions WHERE weekly_session_id = ?`, weeklySessionID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !database.IsNoRows(err) {
		log.Printf("attendance ensure read: %v", err)
		return 0, err
	}
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO attendance_sessions (class_id, session_date, weekly_session_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		classID, sessionDate, weeklySessionID, now, now)
	if err != nil {
		log.Printf("attendance ensure insert: %v", err)
		return 0, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return types.ID(id64), nil
}

func (r *attendanceRepo) attendanceSessionID(ctx context.Context, weeklySessionID types.ID) (types.ID, error) {
	var id types.ID
	err := r.db.QueryRowContext(ctx, `SELECT id FROM attendance_sessions WHERE weekly_session_id = ?`, weeklySessionID).Scan(&id)
	if database.IsNoRows(err) {
		return 0, sql.ErrNoRows
	}
	return id, err
}

func (r *attendanceRepo) ListRecords(ctx context.Context, weeklySessionID types.ID) ([]entity.AttendanceRecord, error) {
	asID, err := r.attendanceSessionID(ctx, weeklySessionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT student_id, status FROM attendance_records WHERE session_id = ? ORDER BY student_id ASC`, asID)
	if err != nil {
		log.Printf("attendance list: %v", err)
		return nil, err
	}
	defer rows.Close()
	var out []entity.AttendanceRecord
	for rows.Next() {
		var rec entity.AttendanceRecord
		if err := rows.Scan(&rec.StudentID, &rec.Status); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *attendanceRepo) UpsertRecord(ctx context.Context, weeklySessionID types.ID, studentID types.ID, status string) error {
	asID, err := r.attendanceSessionID(ctx, weeklySessionID)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO attendance_records (session_id, student_id, status) VALUES (?, ?, ?)
		ON CONFLICT(session_id, student_id) DO UPDATE SET status = excluded.status`,
		asID, studentID, status)
	if err != nil {
		log.Printf("attendance upsert: %v", err)
	}
	return err
}
