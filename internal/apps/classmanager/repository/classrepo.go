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

type ClassRepository interface {
	Create(ctx context.Context, schoolID types.ID, title, grade, room string) (entity.Class, error)
	ListForSchool(ctx context.Context, schoolID types.ID) ([]entity.Class, error)
	ReadByID(ctx context.Context, classID types.ID) (entity.Class, error)
	Update(ctx context.Context, classID, schoolID types.ID, title, grade, room string) (entity.Class, error)
	Delete(ctx context.Context, classID, schoolID types.ID) error
}

type classRepo struct {
	db *sql.DB
}

var _ ClassRepository = (*classRepo)(nil)

func NewClassRepository(db *sql.DB) ClassRepository {
	return &classRepo{db: db}
}

func (r *classRepo) Create(ctx context.Context, schoolID types.ID, title, grade, room string) (entity.Class, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO classes (school_id, title, grade, room, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		schoolID, title, grade, room, now, now)
	if err != nil {
		log.Printf("class create: %v", err)
		return entity.Class{}, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return entity.Class{}, err
	}
	return entity.Class{
		ID:       types.ID(id64),
		SchoolID: schoolID,
		Title:    title,
		Grade:    grade,
		Room:     room,
	}, nil
}

func (r *classRepo) ListForSchool(ctx context.Context, schoolID types.ID) ([]entity.Class, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.school_id, c.title, c.grade, c.room,
			(SELECT COUNT(*) FROM students s WHERE s.class_id = c.id)
		FROM classes c
		WHERE c.school_id = ?
		ORDER BY c.id ASC`,
		schoolID)
	if err != nil {
		log.Printf("class list: %v", err)
		return nil, err
	}
	defer rows.Close()
	var out []entity.Class
	for rows.Next() {
		var c entity.Class
		if err := rows.Scan(&c.ID, &c.SchoolID, &c.Title, &c.Grade, &c.Room, &c.StudentCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *classRepo) ReadByID(ctx context.Context, classID types.ID) (entity.Class, error) {
	return r.scanClass(r.db.QueryRowContext(ctx, `
		SELECT c.id, c.school_id, c.title, c.grade, c.room,
			(SELECT COUNT(*) FROM students s WHERE s.class_id = c.id)
		FROM classes c
		WHERE c.id = ?`,
		classID))
}

func (r *classRepo) Update(ctx context.Context, classID, schoolID types.ID, title, grade, room string) (entity.Class, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		UPDATE classes SET title = ?, grade = ?, room = ?, updated_at = ?
		WHERE id = ? AND school_id = ?`,
		title, grade, room, now, classID, schoolID)
	if err != nil {
		log.Printf("class update: %v", err)
		return entity.Class{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return entity.Class{}, sql.ErrNoRows
	}
	updated, err := r.ReadByID(ctx, classID)
	if err != nil {
		return entity.Class{
			ID:       classID,
			SchoolID: schoolID,
			Title:    title,
			Grade:    grade,
			Room:     room,
		}, nil
	}
	return updated, nil
}

func (r *classRepo) Delete(ctx context.Context, classID, schoolID types.ID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM classes WHERE id = ? AND school_id = ?`, classID, schoolID)
	if err != nil {
		log.Printf("class delete: %v", err)
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *classRepo) scanClass(row *sql.Row) (entity.Class, error) {
	var c entity.Class
	err := row.Scan(&c.ID, &c.SchoolID, &c.Title, &c.Grade, &c.Room, &c.StudentCount)
	if database.IsNoRows(err) {
		return entity.Class{}, nil
	}
	if err != nil {
		return entity.Class{}, err
	}
	return c, nil
}
