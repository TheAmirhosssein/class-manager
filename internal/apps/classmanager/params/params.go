package params

import "github.com/TheAmirhosssein/class-manager/internal/types"

type OwnerActor struct {
	UserID types.ID
}

func (a OwnerActor) Valid() bool {
	return a.UserID.Valid()
}

type CreateSchoolParams struct {
	Actor OwnerActor
	Name  string
}

type UpdateSchoolParams struct {
	Actor    OwnerActor
	SchoolID types.ID
	Name     string
}

type ReadSchoolParams struct {
	Actor    OwnerActor
	SchoolID types.ID
}

type ListSchoolsParams struct {
	Actor OwnerActor
}
