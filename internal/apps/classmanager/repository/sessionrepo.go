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

type SessionRepository interface {
	Create(ctx context.Context, schoolID, classID types.ID, sessionDate string, day int, startTime, endTime, subject string) (entity.Session, error)
	ListForClass(ctx context.Context, classID types.ID) ([]entity.Session, error)
	ListForSchool(ctx context.Context, schoolID types.ID) ([]entity.Session, error)
	ReadByID(ctx context.Context, sessionID types.ID) (entity.Session, error)
	Update(ctx context.Context, sessionID, schoolID, classID types.ID, sessionDate string, day int, startTime, endTime, subject string) (entity.Session, error)
	Delete(ctx context.Context, sessionID, classID types.ID) error
	HasTimeOverlap(ctx context.Context, schoolID types.ID, sessionDate, start, end string, excludeID types.ID) (bool, error)
}

type sessionRepo struct {
	db *sql.DB
}

var _ SessionRepository = (*sessionRepo)(nil)

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepo{db: db}
}

func subjectDB(subject string) any {
	if subject == "" {
		return nil
	}
	return subject
}

func (r *sessionRepo) Create(ctx context.Context, schoolID, classID types.ID, sessionDate string, day int, startTime, endTime, subject string) (entity.Session, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (class_id, school_id, session_date, day, start_time, end_time, subject, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		classID, schoolID, sessionDate, day, startTime, endTime, subjectDB(subject), now, now)
	if err != nil {
		log.Printf("session create: %v", err)
		return entity.Session{}, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return entity.Session{}, err
	}
	return entity.Session{
		ID:          types.ID(id64),
		ClassID:     classID,
		SchoolID:    schoolID,
		SessionDate: sessionDate,
		Day:         day,
		StartTime:   startTime,
		EndTime:     endTime,
		Subject:     subject,
	}, nil
}

const sessionSelect = `
	SELECT id, class_id, school_id, session_date, day, start_time, end_time, subject
	FROM sessions`

func scanSession(sc interface {
	Scan(dest ...any) error
}) (entity.Session, error) {
	var s entity.Session
	var subj sql.NullString
	err := sc.Scan(&s.ID, &s.ClassID, &s.SchoolID, &s.SessionDate, &s.Day, &s.StartTime, &s.EndTime, &subj)
	if err != nil {
		return entity.Session{}, err
	}
	if subj.Valid {
		s.Subject = subj.String
	}
	return s, nil
}

func (r *sessionRepo) ListForClass(ctx context.Context, classID types.ID) ([]entity.Session, error) {
	rows, err := r.db.QueryContext(ctx, sessionSelect+`
		WHERE class_id = ?
		ORDER BY session_date DESC, start_time ASC, id ASC`, classID)
	if err != nil {
		log.Printf("session list class: %v", err)
		return nil, err
	}
	defer rows.Close()
	return scanSessionRows(rows)
}

func (r *sessionRepo) ListForSchool(ctx context.Context, schoolID types.ID) ([]entity.Session, error) {
	rows, err := r.db.QueryContext(ctx, sessionSelect+`
		WHERE school_id = ?
		ORDER BY session_date DESC, start_time ASC, id ASC`, schoolID)
	if err != nil {
		log.Printf("session list school: %v", err)
		return nil, err
	}
	defer rows.Close()
	return scanSessionRows(rows)
}

func scanSessionRows(rows *sql.Rows) ([]entity.Session, error) {
	var out []entity.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *sessionRepo) ReadByID(ctx context.Context, sessionID types.ID) (entity.Session, error) {
	s, err := scanSession(r.db.QueryRowContext(ctx, sessionSelect+` WHERE id = ?`, sessionID))
	if database.IsNoRows(err) {
		return entity.Session{}, nil
	}
	if err != nil {
		return entity.Session{}, err
	}
	return s, nil
}

func (r *sessionRepo) Update(ctx context.Context, sessionID, schoolID, classID types.ID, sessionDate string, day int, startTime, endTime, subject string) (entity.Session, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		UPDATE sessions
		SET session_date = ?, day = ?, start_time = ?, end_time = ?, subject = ?, updated_at = ?
		WHERE id = ? AND class_id = ? AND school_id = ?`,
		sessionDate, day, startTime, endTime, subjectDB(subject), now, sessionID, classID, schoolID)
	if err != nil {
		log.Printf("session update: %v", err)
		return entity.Session{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return entity.Session{}, sql.ErrNoRows
	}
	return r.ReadByID(ctx, sessionID)
}

func (r *sessionRepo) Delete(ctx context.Context, sessionID, classID types.ID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ? AND class_id = ?`, sessionID, classID)
	if err != nil {
		log.Printf("session delete: %v", err)
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *sessionRepo) HasTimeOverlap(ctx context.Context, schoolID types.ID, sessionDate, start, end string, excludeID types.ID) (bool, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id FROM sessions
		WHERE school_id = ? AND session_date = ? AND id != ?
		  AND start_time < ? AND end_time > ?
		LIMIT 1`,
		schoolID, sessionDate, excludeID, end, start)
	var id types.ID
	err := row.Scan(&id)
	if database.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return id.Valid(), nil
}
