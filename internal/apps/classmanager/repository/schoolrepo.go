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

type SchoolRepository interface {
	Create(ctx context.Context, ownerUserID types.ID, name string) (entity.School, error)
	ReadByIDForOwner(ctx context.Context, schoolID, ownerUserID types.ID) (entity.School, error)
	ReadFirstForOwner(ctx context.Context, ownerUserID types.ID) (entity.School, error)
	ListForOwner(ctx context.Context, ownerUserID types.ID) ([]entity.School, error)
	UpdateName(ctx context.Context, schoolID, ownerUserID types.ID, name string) (entity.School, error)
}

type schoolRepo struct {
	db *sql.DB
}

var _ SchoolRepository = (*schoolRepo)(nil)

func NewSchoolRepository(db *sql.DB) SchoolRepository {
	return &schoolRepo{db: db}
}

func (r *schoolRepo) Create(ctx context.Context, ownerUserID types.ID, name string) (entity.School, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO schools (name, owner_user_id, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		name, ownerUserID, now, now)
	if err != nil {
		log.Printf("school create: %v", err)
		return entity.School{}, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return entity.School{}, err
	}
	return entity.School{ID: types.ID(id64), Name: name, OwnerUserID: ownerUserID}, nil
}

func (r *schoolRepo) ReadByIDForOwner(ctx context.Context, schoolID, ownerUserID types.ID) (entity.School, error) {
	return r.scanSchool(r.db.QueryRowContext(ctx, `
		SELECT id, name, owner_user_id FROM schools WHERE id = ? AND owner_user_id = ?`,
		schoolID, ownerUserID))
}

func (r *schoolRepo) ReadFirstForOwner(ctx context.Context, ownerUserID types.ID) (entity.School, error) {
	return r.scanSchool(r.db.QueryRowContext(ctx, `
		SELECT id, name, owner_user_id FROM schools WHERE owner_user_id = ? ORDER BY id ASC LIMIT 1`,
		ownerUserID))
}

func (r *schoolRepo) ListForOwner(ctx context.Context, ownerUserID types.ID) ([]entity.School, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, owner_user_id FROM schools WHERE owner_user_id = ? ORDER BY id ASC`,
		ownerUserID)
	if err != nil {
		log.Printf("school list: %v", err)
		return nil, err
	}
	defer rows.Close()
	var out []entity.School
	for rows.Next() {
		var s entity.School
		var owner sql.NullInt64
		if err := rows.Scan(&s.ID, &s.Name, &owner); err != nil {
			return nil, err
		}
		if owner.Valid {
			s.OwnerUserID = types.ID(owner.Int64)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *schoolRepo) UpdateName(ctx context.Context, schoolID, ownerUserID types.ID, name string) (entity.School, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		UPDATE schools SET name = ?, updated_at = ? WHERE id = ? AND owner_user_id = ?`,
		name, now, schoolID, ownerUserID)
	if err != nil {
		log.Printf("school update: %v", err)
		return entity.School{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return entity.School{}, sql.ErrNoRows
	}
	return entity.School{ID: schoolID, Name: name, OwnerUserID: ownerUserID}, nil
}

func (r *schoolRepo) scanSchool(row *sql.Row) (entity.School, error) {
	var s entity.School
	var owner sql.NullInt64
	err := row.Scan(&s.ID, &s.Name, &owner)
	if database.IsNoRows(err) {
		return entity.School{}, nil
	}
	if err != nil {
		return entity.School{}, err
	}
	if owner.Valid {
		s.OwnerUserID = types.ID(owner.Int64)
	}
	return s, nil
}

func formatSQLiteTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
