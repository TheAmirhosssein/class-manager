package params

import "github.com/TheAmirhosssein/class-manager/internal/types"

type CreateUserParams struct {
	Mobile   string
	Password string
	Name     string
}

type CreateUserUsecaseParams struct {
	ActorAdminID types.ID
	CreateUserParams
}
