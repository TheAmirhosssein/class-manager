package seed

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/enum"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/repository"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/database"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/hash"
)

const (
	TeacherMobile   = "09001110011"
	TeacherPassword = "strong-password"
	TeacherName     = "مدرس نمونه"

	AdminMobile   = "09120000002"
	AdminPassword = "Admin12345"
	AdminName     = "مدیر سامانه"
)

// Create inserts development fixtures when missing. Safe to run repeatedly.
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
	teacherID, err := teacherUserID(ctx, userRepo)
	if err != nil {
		return err
	}
	if err := ensureSampleSchool(ctx, db, teacherID); err != nil {
		return err
	}
	return nil
}

func teacherUserID(ctx context.Context, users repository.UserRepository) (int64, error) {
	u, err := users.ReadByMobile(ctx, TeacherMobile)
	if err != nil {
		return 0, err
	}
	if !u.ID.Valid() {
		return 0, fmt.Errorf("seed teacher missing")
	}
	return int64(u.ID), nil
}

func ensureTeacher(ctx context.Context, users repository.UserRepository, h hash.Hash) error {
	exists, err := users.ExistByMobile(ctx, TeacherMobile)
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
		Mobile:       TeacherMobile,
		PasswordHash: hashVal,
		Name:         TeacherName,
		Role:         enum.RoleTeacher,
	})
	return err
}

func ensureAdmin(ctx context.Context, admins repository.AdminRepository, h hash.Hash) error {
	exists, err := admins.ExistByMobile(ctx, AdminMobile)
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
		Mobile:       AdminMobile,
		PasswordHash: hashVal,
		Name:         AdminName,
	})
	return err
}

func ensureSampleSchool(ctx context.Context, db *sql.DB, ownerUserID int64) error {
	var schoolID int64
	err := db.QueryRowContext(ctx, `
		SELECT id FROM schools WHERE owner_user_id = ? ORDER BY id LIMIT 1`, ownerUserID).Scan(&schoolID)
	if database.IsNoRows(err) {
		res, insErr := db.ExecContext(ctx, `
			INSERT INTO schools (name, owner_user_id) VALUES (?, ?)`, "مدرسه نمونه", ownerUserID)
		if insErr != nil {
			return insErr
		}
		schoolID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	_, _ = db.ExecContext(ctx, `
		UPDATE schools SET owner_user_id = ? WHERE id = ? AND owner_user_id IS NULL`, ownerUserID, schoolID)
	return nil
}
