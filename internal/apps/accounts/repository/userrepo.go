package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/enum"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/validation"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/database"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type UserRepository interface {
	Create(ctx context.Context, u entity.User) (entity.User, error)
	ReadByID(ctx context.Context, id types.ID) (entity.User, error)
	ReadByMobile(ctx context.Context, mobile string) (entity.User, error)
	ExistByMobile(ctx context.Context, mobile string) (bool, error)
}

type userRepo struct {
	db *sql.DB
}

var _ UserRepository = (*userRepo)(nil)

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u entity.User) (entity.User, error) {
	mobile := validation.NormalizeMobile(u.Mobile)
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO users (mobile, password_hash, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		mobile, u.PasswordHash, strings.TrimSpace(u.Name), string(u.Role),
		formatSQLiteTime(now), formatSQLiteTime(now),
	)
	if err != nil {
		log.Printf("accounts user create: %v", err)
		return entity.User{}, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return entity.User{}, err
	}
	u.ID = types.ID(id64)
	u.Mobile = mobile
	u.CreatedAt = now
	u.UpdatedAt = now
	return u, nil
}

func (r *userRepo) ReadByID(ctx context.Context, id types.ID) (entity.User, error) {
	return r.scanOne(r.db.QueryRowContext(ctx, `
		SELECT id, mobile, password_hash, name, role, created_at, updated_at
		FROM users WHERE id = ?`, id))
}

func (r *userRepo) ReadByMobile(ctx context.Context, mobile string) (entity.User, error) {
	mobile = validation.NormalizeMobile(mobile)
	return r.scanOne(r.db.QueryRowContext(ctx, `
		SELECT id, mobile, password_hash, name, role, created_at, updated_at
		FROM users WHERE mobile = ?`, mobile))
}

func (r *userRepo) ExistByMobile(ctx context.Context, mobile string) (bool, error) {
	mobile = validation.NormalizeMobile(mobile)
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM users WHERE mobile = ? LIMIT 1`, mobile).Scan(&n)
	if database.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		log.Printf("accounts user exist by mobile: %v", err)
		return false, err
	}
	return true, nil
}

func (r *userRepo) scanOne(row *sql.Row) (entity.User, error) {
	var u entity.User
	var id int64
	var role string
	var createdAt, updatedAt string
	err := row.Scan(&id, &u.Mobile, &u.PasswordHash, &u.Name, &role, &createdAt, &updatedAt)
	if database.IsNoRows(err) {
		return entity.User{}, nil
	}
	if err != nil {
		log.Printf("accounts user scan: %v", err)
		return entity.User{}, err
	}
	u.ID = types.ID(id)
	u.Role = enum.UserRole(role)
	var parseErr error
	u.CreatedAt, parseErr = parseSQLiteTime(createdAt)
	if parseErr != nil {
		return entity.User{}, fmt.Errorf("parse created_at: %w", parseErr)
	}
	u.UpdatedAt, parseErr = parseSQLiteTime(updatedAt)
	if parseErr != nil {
		return entity.User{}, fmt.Errorf("parse updated_at: %w", parseErr)
	}
	return u, nil
}
