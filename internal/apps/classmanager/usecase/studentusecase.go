package usecase

import (
	"context"
	"database/sql"
	"strings"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/validation"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type StudentUsecase struct {
	students repository.StudentRepository
	classes  repository.ClassRepository
	schools  repository.SchoolRepository
}

func NewStudentUsecase(students repository.StudentRepository, classes repository.ClassRepository, schools repository.SchoolRepository) *StudentUsecase {
	return &StudentUsecase{students: students, classes: classes, schools: schools}
}

func (u *StudentUsecase) classForOwner(ctx context.Context, actor params.OwnerActor, classID types.ID) (entity.Class, messages.Message) {
	if !classID.Valid() {
		return entity.Class{}, message.InvalidClassInput
	}
	if !actor.Valid() {
		return entity.Class{}, message.Forbidden
	}
	class, err := u.classes.ReadByID(ctx, classID)
	if err != nil {
		log.Printf("student class read: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	if !class.ID.Valid() {
		return entity.Class{}, message.ClassNotFound
	}
	school, err := u.schools.ReadByIDForOwner(ctx, class.SchoolID, actor.UserID)
	if err != nil {
		log.Printf("student owner school check: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return entity.Class{}, message.Forbidden
	}
	return class, messages.OK()
}

func (u *StudentUsecase) studentForOwner(ctx context.Context, actor params.OwnerActor, studentID types.ID) (entity.Student, messages.Message) {
	if !studentID.Valid() {
		return entity.Student{}, message.InvalidStudentInput
	}
	if !actor.Valid() {
		return entity.Student{}, message.Forbidden
	}
	student, err := u.students.ReadByID(ctx, studentID)
	if err != nil {
		log.Printf("student read: %v", err)
		return entity.Student{}, messages.SharedServerError()
	}
	if !student.ID.Valid() {
		return entity.Student{}, message.StudentNotFound
	}
	if _, msg := u.classForOwner(ctx, actor, student.ClassID); msg.Error != nil {
		return entity.Student{}, msg
	}
	return student, messages.OK()
}

type normalizedStudentFields struct {
	firstName     string
	lastName      string
	nationalCode  string
	fatherName    string
	studentNumber string
}

func (u *StudentUsecase) normalizeFields(firstName, lastName, nationalCode, fatherName, studentNumber string) (normalizedStudentFields, messages.Message) {
	firstName = validation.NormalizeStudentFirstName(firstName)
	lastName = validation.NormalizeStudentLastName(lastName)
	nationalCode = validation.NormalizeNationalCode(nationalCode)
	fatherName = validation.NormalizeStudentFatherName(fatherName)
	studentNumber = validation.NormalizeStudentNumber(studentNumber)

	if !validation.ValidStudentFirstName(firstName) ||
		!validation.ValidStudentLastName(lastName) ||
		!validation.ValidNationalCode(nationalCode) ||
		!validation.ValidStudentFatherName(fatherName) ||
		!validation.ValidStudentNumber(studentNumber) {
		return normalizedStudentFields{}, message.InvalidStudentInput
	}
	return normalizedStudentFields{
		firstName:     firstName,
		lastName:      lastName,
		nationalCode:  nationalCode,
		fatherName:    fatherName,
		studentNumber: studentNumber,
	}, messages.OK()
}

func (u *StudentUsecase) ensureNationalCodeFree(ctx context.Context, nationalCode string, exclude types.ID) messages.Message {
	if nationalCode == "" {
		return messages.OK()
	}
	taken, err := u.students.NationalCodeTaken(ctx, nationalCode, exclude)
	if err != nil {
		log.Printf("student national code check: %v", err)
		return messages.SharedServerError()
	}
	if taken {
		return message.NationalCodeTaken
	}
	return messages.OK()
}

func (u *StudentUsecase) List(ctx context.Context, p params.ListStudentsParams) ([]entity.Student, messages.Message) {
	if _, msg := u.classForOwner(ctx, p.Actor, p.ClassID); msg.Error != nil {
		return nil, msg
	}
	list, err := u.students.ListForClass(ctx, p.ClassID)
	if err != nil {
		log.Printf("student list usecase: %v", err)
		return nil, messages.SharedServerError()
	}
	return list, message.StudentRead
}

func (u *StudentUsecase) Read(ctx context.Context, p params.ReadStudentParams) (entity.Student, messages.Message) {
	return u.studentForOwner(ctx, p.Actor, p.StudentID)
}

func (u *StudentUsecase) Create(ctx context.Context, p params.CreateStudentParams) (entity.Student, messages.Message) {
	class, msg := u.classForOwner(ctx, p.Actor, p.ClassID)
	if msg.Error != nil {
		return entity.Student{}, msg
	}
	fields, msg := u.normalizeFields(p.FirstName, p.LastName, p.NationalCode, p.FatherName, p.StudentNumber)
	if msg.Error != nil {
		return entity.Student{}, msg
	}
	if msg = u.ensureNationalCodeFree(ctx, fields.nationalCode, 0); msg.Error != nil {
		return entity.Student{}, msg
	}
	student, err := u.students.Create(ctx, class.ID, fields.firstName, fields.lastName, fields.nationalCode, fields.fatherName, fields.studentNumber)
	if err != nil {
		if isSQLiteUniqueNationalCode(err) {
			return entity.Student{}, message.NationalCodeTaken
		}
		log.Printf("student create usecase: %v", err)
		return entity.Student{}, messages.SharedServerError()
	}
	return student, message.StudentCreated
}

func (u *StudentUsecase) Update(ctx context.Context, p params.UpdateStudentParams) (entity.Student, messages.Message) {
	student, msg := u.studentForOwner(ctx, p.Actor, p.StudentID)
	if msg.Error != nil {
		return entity.Student{}, msg
	}
	if p.ClassID.Valid() && p.ClassID != student.ClassID {
		return entity.Student{}, message.InvalidStudentInput
	}
	fields, msg := u.normalizeFields(p.FirstName, p.LastName, p.NationalCode, p.FatherName, p.StudentNumber)
	if msg.Error != nil {
		return entity.Student{}, msg
	}
	if msg = u.ensureNationalCodeFree(ctx, fields.nationalCode, student.ID); msg.Error != nil {
		return entity.Student{}, msg
	}
	updated, err := u.students.Update(ctx, student.ID, student.ClassID, fields.firstName, fields.lastName, fields.nationalCode, fields.fatherName, fields.studentNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.Student{}, message.StudentNotFound
		}
		if isSQLiteUniqueNationalCode(err) {
			return entity.Student{}, message.NationalCodeTaken
		}
		log.Printf("student update usecase: %v", err)
		return entity.Student{}, messages.SharedServerError()
	}
	return updated, message.StudentUpdated
}

func (u *StudentUsecase) Delete(ctx context.Context, p params.DeleteStudentParams) messages.Message {
	student, msg := u.studentForOwner(ctx, p.Actor, p.StudentID)
	if msg.Error != nil {
		return msg
	}
	err := u.students.Delete(ctx, student.ID, student.ClassID)
	if err != nil {
		if err == sql.ErrNoRows {
			return message.StudentNotFound
		}
		log.Printf("student delete usecase: %v", err)
		return messages.SharedServerError()
	}
	return message.StudentDeleted
}

func isSQLiteUniqueNationalCode(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") && strings.Contains(msg, "national_code")
}
