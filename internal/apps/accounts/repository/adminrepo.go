package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/validation"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/database"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type AdminRepository interface {
	Create(ctx context.Context, a entity.Admin) (entity.Admin, error)
	ReadByID(ctx context.Context, id types.ID) (entity.Admin, error)
	ReadByMobile(ctx context.Context, mobile string) (entity.Admin, error)
	ExistByMobile(ctx context.Context, mobile string) (bool, error)
}

type adminRepo struct {
	db *sql.DB
}

var _ AdminRepository = (*adminRepo)(nil)

func NewAdminRepository(db *sql.DB) AdminRepository {
	return &adminRepo{db: db}
}

func (r *adminRepo) Create(ctx context.Context, a entity.Admin) (entity.Admin, error) {
	mobile := validation.NormalizeMobile(a.Mobile)
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO admins (mobile, password_hash, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		mobile, a.PasswordHash, strings.TrimSpace(a.Name),
		formatSQLiteTime(now), formatSQLiteTime(now),
	)
	if err != nil {
		log.Printf("accounts admin create: %v", err)
		return entity.Admin{}, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return entity.Admin{}, err
	}
	a.ID = types.ID(id64)
	a.Mobile = mobile
	a.CreatedAt = now
	a.UpdatedAt = now
	return a, nil
}

func (r *adminRepo) ReadByID(ctx context.Context, id types.ID) (entity.Admin, error) {
	return r.scanOne(r.db.QueryRowContext(ctx, `
		SELECT id, mobile, password_hash, name, created_at, updated_at
		FROM admins WHERE id = ?`, id))
}

func (r *adminRepo) ReadByMobile(ctx context.Context, mobile string) (entity.Admin, error) {
	mobile = validation.NormalizeMobile(mobile)
	return r.scanOne(r.db.QueryRowContext(ctx, `
		SELECT id, mobile, password_hash, name, created_at, updated_at
		FROM admins WHERE mobile = ?`, mobile))
}

func (r *adminRepo) ExistByMobile(ctx context.Context, mobile string) (bool, error) {
	mobile = validation.NormalizeMobile(mobile)
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM admins WHERE mobile = ? LIMIT 1`, mobile).Scan(&n)
	if database.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		log.Printf("accounts admin exist by mobile: %v", err)
		return false, err
	}
	return true, nil
}

func (r *adminRepo) scanOne(row *sql.Row) (entity.Admin, error) {
	var a entity.Admin
	var id int64
	var createdAt, updatedAt string
	err := row.Scan(&id, &a.Mobile, &a.PasswordHash, &a.Name, &createdAt, &updatedAt)
	if database.IsNoRows(err) {
		return entity.Admin{}, nil
	}
	if err != nil {
		log.Printf("accounts admin scan: %v", err)
		return entity.Admin{}, err
	}
	a.ID = types.ID(id)
	var parseErr error
	a.CreatedAt, parseErr = parseSQLiteTime(createdAt)
	if parseErr != nil {
		return entity.Admin{}, fmt.Errorf("parse created_at: %w", parseErr)
	}
	a.UpdatedAt, parseErr = parseSQLiteTime(updatedAt)
	if parseErr != nil {
		return entity.Admin{}, fmt.Errorf("parse updated_at: %w", parseErr)
	}
	return a, nil
}
