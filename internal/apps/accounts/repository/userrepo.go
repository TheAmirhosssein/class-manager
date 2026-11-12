package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/enum"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/database"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type UserRepository interface {
	Create(ctx context.Context, u entity.User) (entity.User, error)
	ReadByID(ctx context.Context, id types.ID) (entity.User, error)
	ReadByEmail(ctx context.Context, email string) (entity.User, error)
	ExistByEmail(ctx context.Context, email string) (bool, error)
}

type userRepo struct {
	db *sql.DB
}

var _ UserRepository = (*userRepo)(nil)

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u entity.User) (entity.User, error) {
	email := strings.TrimSpace(u.Email)
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		email, u.PasswordHash, strings.TrimSpace(u.Name), string(u.Role),
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
	u.Email = email
	u.CreatedAt = now
	u.UpdatedAt = now
	return u, nil
}

func (r *userRepo) ReadByID(ctx context.Context, id types.ID) (entity.User, error) {
	return r.scanOne(r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users WHERE id = ?`, id))
}

func (r *userRepo) ReadByEmail(ctx context.Context, email string) (entity.User, error) {
	email = strings.TrimSpace(email)
	return r.scanOne(r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users WHERE email = ? COLLATE NOCASE`, email))
}

func (r *userRepo) ExistByEmail(ctx context.Context, email string) (bool, error) {
	email = strings.TrimSpace(email)
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM users WHERE email = ? COLLATE NOCASE LIMIT 1`, email).Scan(&n)
	if database.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		log.Printf("accounts user exist by email: %v", err)
		return false, err
	}
	return true, nil
}

func (r *userRepo) scanOne(row *sql.Row) (entity.User, error) {
	var u entity.User
	var id int64
	var role string
	var createdAt, updatedAt string
	err := row.Scan(&id, &u.Email, &u.PasswordHash, &u.Name, &role, &createdAt, &updatedAt)
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
