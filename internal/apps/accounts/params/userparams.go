package params

import "github.com/TheAmirhosssein/class-manager/internal/types"

type CreateUserParams struct {
	Email    string
	Password string
	Name     string
}

type CreateUserUsecaseParams struct {
	ActorAdminID types.ID
	CreateUserParams
}
