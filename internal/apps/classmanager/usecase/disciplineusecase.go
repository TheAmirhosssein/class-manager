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

type DisciplineUsecase struct {
	records  repository.DisciplineRepository
	students repository.StudentRepository
	classes  repository.ClassRepository
	schools  repository.SchoolRepository
}

func NewDisciplineUsecase(records repository.DisciplineRepository, students repository.StudentRepository, classes repository.ClassRepository, schools repository.SchoolRepository) *DisciplineUsecase {
	return &DisciplineUsecase{records: records, students: students, classes: classes, schools: schools}
}

func (u *DisciplineUsecase) resolveSchoolID(ctx context.Context, actor params.OwnerActor, schoolID types.ID) (types.ID, messages.Message) {
	if !actor.Valid() {
		return 0, message.Forbidden
	}
	if schoolID.Valid() {
		school, err := u.schools.ReadByIDForOwner(ctx, schoolID, actor.UserID)
		if err != nil {
			log.Printf("discipline school owner check: %v", err)
			return 0, messages.SharedServerError()
		}
		if !school.ID.Valid() {
			return 0, message.Forbidden
		}
		return school.ID, messages.OK()
	}
	school, err := u.schools.ReadFirstForOwner(ctx, actor.UserID)
	if err != nil {
		log.Printf("discipline resolve first school: %v", err)
		return 0, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return 0, message.NoSchool
	}
	return school.ID, messages.OK()
}

func (u *DisciplineUsecase) classForOwner(ctx context.Context, actor params.OwnerActor, classID types.ID) (entity.Class, messages.Message) {
	if !classID.Valid() {
		return entity.Class{}, message.InvalidClassInput
	}
	if !actor.Valid() {
		return entity.Class{}, message.Forbidden
	}
	class, err := u.classes.ReadByID(ctx, classID)
	if err != nil {
		log.Printf("discipline class read: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	if !class.ID.Valid() {
		return entity.Class{}, message.ClassNotFound
	}
	school, err := u.schools.ReadByIDForOwner(ctx, class.SchoolID, actor.UserID)
	if err != nil {
		log.Printf("discipline owner school check: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return entity.Class{}, message.Forbidden
	}
	return class, messages.OK()
}

func studentDisplayName(s entity.Student) string {
	return strings.TrimSpace(strings.TrimSpace(s.FirstName + " " + s.LastName))
}

func (u *DisciplineUsecase) recordForOwner(ctx context.Context, actor params.OwnerActor, recordID types.ID) (entity.DisciplineRecord, messages.Message) {
	if !recordID.Valid() {
		return entity.DisciplineRecord{}, message.InvalidDisciplineInput
	}
	record, err := u.records.ReadByID(ctx, recordID)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.DisciplineRecord{}, message.DisciplineNotFound
		}
		log.Printf("discipline read: %v", err)
		return entity.DisciplineRecord{}, messages.SharedServerError()
	}
	if _, msg := u.classForOwner(ctx, actor, record.ClassID); msg.Error != nil {
		return entity.DisciplineRecord{}, msg
	}
	return record, messages.OK()
}

func (u *DisciplineUsecase) studentInClass(ctx context.Context, actor params.OwnerActor, classID, studentID types.ID) (entity.Student, messages.Message) {
	student, err := u.students.ReadByID(ctx, studentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.Student{}, message.StudentNotFound
		}
		log.Printf("discipline student read: %v", err)
		return entity.Student{}, messages.SharedServerError()
	}
	if !student.ID.Valid() || student.ClassID != classID {
		return entity.Student{}, message.StudentNotFound
	}
	if _, msg := u.classForOwner(ctx, actor, classID); msg.Error != nil {
		return entity.Student{}, msg
	}
	return student, messages.OK()
}

type normalizedDisciplineFields struct {
	recordType string
	note       string
	recordDate string
}

func normalizeDiscipline(recordType, note, recordDate string) (normalizedDisciplineFields, messages.Message) {
	recordType = validation.NormalizeDisciplineType(recordType)
	note = validation.NormalizeDisciplineNote(note)
	recordDate = validation.NormalizeDisciplineDate(recordDate)
	if !validation.ValidDisciplineType(recordType) || !validation.ValidDisciplineNote(note) || !validation.ValidDisciplineDate(recordDate) {
		return normalizedDisciplineFields{}, message.InvalidDisciplineInput
	}
	return normalizedDisciplineFields{recordType: recordType, note: note, recordDate: recordDate}, messages.OK()
}

func (u *DisciplineUsecase) List(ctx context.Context, p params.ListDisciplineParams) ([]entity.DisciplineRecord, messages.Message) {
	schoolID, msg := u.resolveSchoolID(ctx, p.Actor, p.SchoolID)
	if msg.Error != nil {
		return nil, msg
	}
	if p.ClassID.Valid() {
		if _, msg := u.classForOwner(ctx, p.Actor, p.ClassID); msg.Error != nil {
			return nil, msg
		}
	}
	if p.StudentID.Valid() {
		student, err := u.students.ReadByID(ctx, p.StudentID)
		if err != nil || !student.ID.Valid() {
			return []entity.DisciplineRecord{}, message.DisciplineRead
		}
		if _, msg := u.classForOwner(ctx, p.Actor, student.ClassID); msg.Error != nil {
			return nil, msg
		}
	}
	list, err := u.records.ListForSchool(ctx, schoolID, p.ClassID, p.StudentID)
	if err != nil {
		log.Printf("discipline list usecase: %v", err)
		return nil, messages.SharedServerError()
	}
	return list, message.DisciplineRead
}

func (u *DisciplineUsecase) Create(ctx context.Context, p params.CreateDisciplineParams) (entity.DisciplineRecord, messages.Message) {
	class, msg := u.classForOwner(ctx, p.Actor, p.ClassID)
	if msg.Error != nil {
		return entity.DisciplineRecord{}, msg
	}
	student, msg := u.studentInClass(ctx, p.Actor, class.ID, p.StudentID)
	if msg.Error != nil {
		return entity.DisciplineRecord{}, msg
	}
	fields, msg := normalizeDiscipline(p.Type, p.Note, p.RecordDate)
	if msg.Error != nil {
		return entity.DisciplineRecord{}, msg
	}
	name := studentDisplayName(student)
	record, err := u.records.Create(ctx, class.ID, student.ID, name, fields.recordType, fields.note, fields.recordDate)
	if err != nil {
		log.Printf("discipline create usecase: %v", err)
		return entity.DisciplineRecord{}, messages.SharedServerError()
	}
	return record, message.DisciplineCreated
}

func (u *DisciplineUsecase) Update(ctx context.Context, p params.UpdateDisciplineParams) (entity.DisciplineRecord, messages.Message) {
	existing, msg := u.recordForOwner(ctx, p.Actor, p.RecordID)
	if msg.Error != nil {
		return entity.DisciplineRecord{}, msg
	}
	classID := existing.ClassID
	if p.ClassID.Valid() {
		classID = p.ClassID
	}
	class, msg := u.classForOwner(ctx, p.Actor, classID)
	if msg.Error != nil {
		return entity.DisciplineRecord{}, msg
	}
	studentID := existing.StudentID
	if p.StudentID.Valid() {
		studentID = p.StudentID
	}
	student, msg := u.studentInClass(ctx, p.Actor, class.ID, studentID)
	if msg.Error != nil {
		return entity.DisciplineRecord{}, msg
	}
	fields, msg := normalizeDiscipline(p.Type, p.Note, p.RecordDate)
	if msg.Error != nil {
		return entity.DisciplineRecord{}, msg
	}
	name := studentDisplayName(student)
	record, err := u.records.Update(ctx, existing.ID, class.ID, student.ID, name, fields.recordType, fields.note, fields.recordDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.DisciplineRecord{}, message.DisciplineNotFound
		}
		log.Printf("discipline update usecase: %v", err)
		return entity.DisciplineRecord{}, messages.SharedServerError()
	}
	return record, message.DisciplineUpdated
}

func (u *DisciplineUsecase) Delete(ctx context.Context, p params.DeleteDisciplineParams) messages.Message {
	if _, msg := u.recordForOwner(ctx, p.Actor, p.RecordID); msg.Error != nil {
		return msg
	}
	if err := u.records.Delete(ctx, p.RecordID); err != nil {
		if err == sql.ErrNoRows {
			return message.DisciplineNotFound
		}
		log.Printf("discipline delete usecase: %v", err)
		return messages.SharedServerError()
	}
	return message.DisciplineDeleted
}
