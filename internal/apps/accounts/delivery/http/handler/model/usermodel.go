package model

import (
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/enum"
)

type CreateUserBody struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type CreatedUserJSON struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Mobile string `json:"mobile"`
}

func NewCreatedUserJSON(u entity.User) CreatedUserJSON {
	return CreatedUserJSON{
		ID:     strconv.FormatInt(int64(u.ID), 10),
		Name:   u.Name,
		Role:   string(enum.RoleTeacher),
		Mobile: u.Mobile,
	}
}
