package seed

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/enum"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/repository"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/hash"
)

const (
	TeacherEmail    = "teacher@classino.local"
	TeacherPassword = "Teacher123"
	TeacherName     = "مدرس نمونه"

	AdminEmail    = "admin@classino.local"
	AdminPassword = "Admin12345"
	AdminName     = "مدیر سامانه"
)

// Create inserts demo accounts when missing. Safe to run repeatedly.
func Create(ctx context.Context, db *sql.DB) error {
	h := hash.New()
	userRepo := repository.NewUserRepository(db)
	adminRepo := repository.NewAdminRepository(db)

	if err := ensureTeacher(ctx, userRepo, h); err != nil {
		return err
	}
	if err := ensureAdmin(ctx, adminRepo, h); err != nil {
		return err
	}
	if err := ensureSampleSchool(ctx, db); err != nil {
		return err
	}
	return nil
}

func ensureTeacher(ctx context.Context, users repository.UserRepository, h hash.Hash) error {
	exists, err := users.ExistByEmail(ctx, TeacherEmail)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	hashVal, err := h.Encrypt(TeacherPassword)
	if err != nil {
		return fmt.Errorf("hash teacher password: %w", err)
	}
	_, err = users.Create(ctx, entity.User{
		Email:        TeacherEmail,
		PasswordHash: hashVal,
		Name:         TeacherName,
		Role:         enum.RoleTeacher,
	})
	return err
}

func ensureAdmin(ctx context.Context, admins repository.AdminRepository, h hash.Hash) error {
	exists, err := admins.ExistByEmail(ctx, AdminEmail)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	hashVal, err := h.Encrypt(AdminPassword)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	_, err = admins.Create(ctx, entity.Admin{
		Email:        AdminEmail,
		PasswordHash: hashVal,
		Name:         AdminName,
	})
	return err
}

func ensureSampleSchool(ctx context.Context, db *sql.DB) error {
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schools`).Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = db.ExecContext(ctx, `INSERT INTO schools (name) VALUES (?)`, "مدرسه نمونه")
	return err
}
