package usecase

import (
	"context"
	"database/sql"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/validation"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/database"
	"github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type SchoolUsecase struct {
	schools repository.SchoolRepository
}

func NewSchoolUsecase(schools repository.SchoolRepository) *SchoolUsecase {
	return &SchoolUsecase{schools: schools}
}

func (u *SchoolUsecase) requireActor(actor params.OwnerActor) messages.Message {
	if !actor.Valid() {
		return message.Forbidden
	}
	return messages.OK()
}

func (u *SchoolUsecase) resolveSchool(ctx context.Context, p params.ReadSchoolParams) (entity.School, messages.Message) {
	if msg := u.requireActor(p.Actor); msg.Error != nil {
		return entity.School{}, msg
	}
	if p.SchoolID.Valid() {
		school, err := u.schools.ReadByIDForOwner(ctx, p.SchoolID, p.Actor.UserID)
		if err != nil {
			log.Printf("school read by id: %v", err)
			return entity.School{}, messages.SharedServerError()
		}
		if !school.ID.Valid() {
			return entity.School{}, message.Forbidden
		}
		return school, message.SchoolRead
	}
	school, err := u.schools.ReadFirstForOwner(ctx, p.Actor.UserID)
	if err != nil {
		log.Printf("school read first: %v", err)
		return entity.School{}, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return entity.School{}, message.NoSchool
	}
	return school, message.SchoolRead
}

func (u *SchoolUsecase) List(ctx context.Context, p params.ListSchoolsParams) ([]entity.School, messages.Message) {
	if msg := u.requireActor(p.Actor); msg.Error != nil {
		return nil, msg
	}
	list, err := u.schools.ListForOwner(ctx, p.Actor.UserID)
	if err != nil {
		log.Printf("school list usecase: %v", err)
		return nil, messages.SharedServerError()
	}
	return list, message.SchoolRead
}

func (u *SchoolUsecase) Create(ctx context.Context, p params.CreateSchoolParams) (entity.School, messages.Message) {
	if msg := u.requireActor(p.Actor); msg.Error != nil {
		return entity.School{}, msg
	}
	name := validation.NormalizeSchoolName(p.Name)
	if !validation.ValidSchoolName(name) {
		return entity.School{}, message.InvalidInput
	}
	taken, err := u.schools.NameExistsForOwner(ctx, p.Actor.UserID, name, 0)
	if err != nil {
		log.Printf("school create name check: %v", err)
		return entity.School{}, messages.SharedServerError()
	}
	if taken {
		return entity.School{}, message.NameTaken
	}
	school, err := u.schools.Create(ctx, p.Actor.UserID, name)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return entity.School{}, message.NameTaken
		}
		log.Printf("school create usecase: %v", err)
		return entity.School{}, messages.SharedServerError()
	}
	return school, message.SchoolCreated
}

func (u *SchoolUsecase) Read(ctx context.Context, p params.ReadSchoolParams) (entity.School, messages.Message) {
	return u.resolveSchool(ctx, p)
}

func (u *SchoolUsecase) UpdateName(ctx context.Context, p params.UpdateSchoolParams) (entity.School, messages.Message) {
	if msg := u.requireActor(p.Actor); msg.Error != nil {
		return entity.School{}, msg
	}
	name := validation.NormalizeSchoolName(p.Name)
	if !validation.ValidSchoolName(name) {
		return entity.School{}, message.InvalidInput
	}
	schoolID := p.SchoolID
	if !schoolID.Valid() {
		current, msg := u.resolveSchool(ctx, params.ReadSchoolParams{Actor: p.Actor})
		if msg.Error != nil {
			return entity.School{}, msg
		}
		schoolID = current.ID
	} else {
		owned, err := u.schools.ReadByIDForOwner(ctx, schoolID, p.Actor.UserID)
		if err != nil {
			log.Printf("school update owner check: %v", err)
			return entity.School{}, messages.SharedServerError()
		}
		if !owned.ID.Valid() {
			return entity.School{}, message.Forbidden
		}
	}
	taken, err := u.schools.NameExistsForOwner(ctx, p.Actor.UserID, name, schoolID)
	if err != nil {
		log.Printf("school update name check: %v", err)
		return entity.School{}, messages.SharedServerError()
	}
	if taken {
		return entity.School{}, message.NameTaken
	}
	updated, err := u.schools.UpdateName(ctx, schoolID, p.Actor.UserID, name)
	if err != nil {
		if database.IsNoRows(err) || err == sql.ErrNoRows {
			return entity.School{}, message.Forbidden
		}
		if database.IsUniqueViolation(err) {
			return entity.School{}, message.NameTaken
		}
		log.Printf("school update usecase: %v", err)
		return entity.School{}, messages.SharedServerError()
	}
	return updated, message.SchoolUpdated
}
