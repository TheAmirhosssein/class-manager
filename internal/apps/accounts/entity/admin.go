package entity

import "github.com/TheAmirhosssein/class-manager/internal/utils/base"

type Admin struct {
	base.Entity
	Email        string
	PasswordHash string
	Name         string
}
