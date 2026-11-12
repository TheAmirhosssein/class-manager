package usecase

import (
	"context"
	"database/sql"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/validation"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type ClassUsecase struct {
	classes repository.ClassRepository
	schools repository.SchoolRepository
}

func NewClassUsecase(classes repository.ClassRepository, schools repository.SchoolRepository) *ClassUsecase {
	return &ClassUsecase{classes: classes, schools: schools}
}

func (u *ClassUsecase) requireActor(actor params.OwnerActor) messages.Message {
	if !actor.Valid() {
		return message.Forbidden
	}
	return messages.OK()
}

func (u *ClassUsecase) resolveSchoolID(ctx context.Context, actor params.OwnerActor, schoolID types.ID) (types.ID, messages.Message) {
	if msg := u.requireActor(actor); msg.Error != nil {
		return 0, msg
	}
	if schoolID.Valid() {
		school, err := u.schools.ReadByIDForOwner(ctx, schoolID, actor.UserID)
		if err != nil {
			log.Printf("class school owner check: %v", err)
			return 0, messages.SharedServerError()
		}
		if !school.ID.Valid() {
			return 0, message.Forbidden
		}
		return school.ID, messages.OK()
	}
	school, err := u.schools.ReadFirstForOwner(ctx, actor.UserID)
	if err != nil {
		log.Printf("class resolve first school: %v", err)
		return 0, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return 0, message.NoSchool
	}
	return school.ID, messages.OK()
}

func (u *ClassUsecase) classForOwner(ctx context.Context, actor params.OwnerActor, classID types.ID) (entity.Class, messages.Message) {
	if !classID.Valid() {
		return entity.Class{}, message.InvalidClassInput
	}
	if msg := u.requireActor(actor); msg.Error != nil {
		return entity.Class{}, msg
	}
	class, err := u.classes.ReadByID(ctx, classID)
	if err != nil {
		log.Printf("class read: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	if !class.ID.Valid() {
		return entity.Class{}, message.ClassNotFound
	}
	school, err := u.schools.ReadByIDForOwner(ctx, class.SchoolID, actor.UserID)
	if err != nil {
		log.Printf("class owner school check: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return entity.Class{}, message.Forbidden
	}
	return class, messages.OK()
}

func (u *ClassUsecase) normalizeFields(title, grade, room string) (string, string, string, messages.Message) {
	title = validation.NormalizeClassTitle(title)
	grade = validation.NormalizeClassGrade(grade)
	room = validation.NormalizeClassRoom(room)
	if !validation.ValidClassTitle(title) || !validation.ValidClassGrade(grade) || !validation.ValidClassRoom(room) {
		return "", "", "", message.InvalidClassInput
	}
	return title, grade, room, messages.OK()
}

func (u *ClassUsecase) List(ctx context.Context, p params.ListClassesParams) ([]entity.Class, messages.Message) {
	schoolID, msg := u.resolveSchoolID(ctx, p.Actor, p.SchoolID)
	if msg.Error != nil {
		return nil, msg
	}
	list, err := u.classes.ListForSchool(ctx, schoolID)
	if err != nil {
		log.Printf("class list usecase: %v", err)
		return nil, messages.SharedServerError()
	}
	return list, message.ClassRead
}

func (u *ClassUsecase) Read(ctx context.Context, p params.ReadClassParams) (entity.Class, messages.Message) {
	return u.classForOwner(ctx, p.Actor, p.ClassID)
}

func (u *ClassUsecase) Create(ctx context.Context, p params.CreateClassParams) (entity.Class, messages.Message) {
	schoolID, msg := u.resolveSchoolID(ctx, p.Actor, p.SchoolID)
	if msg.Error != nil {
		return entity.Class{}, msg
	}
	title, grade, room, msg := u.normalizeFields(p.Title, p.Grade, p.Room)
	if msg.Error != nil {
		return entity.Class{}, msg
	}
	class, err := u.classes.Create(ctx, schoolID, title, grade, room)
	if err != nil {
		log.Printf("class create usecase: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	return class, message.ClassCreated
}

func (u *ClassUsecase) Update(ctx context.Context, p params.UpdateClassParams) (entity.Class, messages.Message) {
	class, msg := u.classForOwner(ctx, p.Actor, p.ClassID)
	if msg.Error != nil {
		return entity.Class{}, msg
	}
	title, grade, room, msg := u.normalizeFields(p.Title, p.Grade, p.Room)
	if msg.Error != nil {
		return entity.Class{}, msg
	}
	if p.SchoolID.Valid() && p.SchoolID != class.SchoolID {
		return entity.Class{}, message.Forbidden
	}
	updated, err := u.classes.Update(ctx, class.ID, class.SchoolID, title, grade, room)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.Class{}, message.ClassNotFound
		}
		log.Printf("class update usecase: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	return updated, message.ClassUpdated
}

func (u *ClassUsecase) Delete(ctx context.Context, p params.DeleteClassParams) messages.Message {
	class, msg := u.classForOwner(ctx, p.Actor, p.ClassID)
	if msg.Error != nil {
		return msg
	}
	err := u.classes.Delete(ctx, class.ID, class.SchoolID)
	if err != nil {
		if err == sql.ErrNoRows {
			return message.ClassNotFound
		}
		log.Printf("class delete usecase: %v", err)
		return messages.SharedServerError()
	}
	return message.ClassDeleted
}
