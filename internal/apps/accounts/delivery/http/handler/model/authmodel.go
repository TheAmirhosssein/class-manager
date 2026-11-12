package model

import (
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/enum"
)

type LoginModel struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserJSON struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	Email string `json:"email,omitempty"`
}

type LoginResponse struct {
	User  UserJSON `json:"user"`
	Token string   `json:"token"`
}

type SessionResponse struct {
	User UserJSON `json:"user"`
}

func NewUserJSON(u entity.User) UserJSON {
	return UserJSON{
		ID:    strconv.FormatInt(int64(u.ID), 10),
		Name:  u.Name,
		Role:  string(enum.RoleTeacher),
		Email: u.Email,
	}
}
