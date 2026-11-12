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

type CreateClassParams struct {
	Actor    OwnerActor
	SchoolID types.ID
	Title    string
	Grade    string
	Room     string
}

type UpdateClassParams struct {
	Actor    OwnerActor
	ClassID  types.ID
	SchoolID types.ID
	Title    string
	Grade    string
	Room     string
}

type ReadClassParams struct {
	Actor   OwnerActor
	ClassID types.ID
}

type ListClassesParams struct {
	Actor    OwnerActor
	SchoolID types.ID
}

type DeleteClassParams struct {
	Actor   OwnerActor
	ClassID types.ID
}
