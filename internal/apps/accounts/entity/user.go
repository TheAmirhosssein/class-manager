package entity

import (
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/enum"
	"github.com/TheAmirhosssein/class-manager/internal/utils/base"
)

type User struct {
	base.Entity
	Email        string
	PasswordHash string
	Name         string
	Role         enum.UserRole
}

func (u User) Teacher() bool {
	return u.Role == enum.RoleTeacher
}
