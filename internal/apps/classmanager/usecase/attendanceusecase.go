package usecase

import (
	"context"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/validation"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type AttendanceUsecase struct {
	attendance repository.AttendanceRepository
	sessions   repository.SessionRepository
	students   repository.StudentRepository
	classes    repository.ClassRepository
	schools    repository.SchoolRepository
}

func NewAttendanceUsecase(
	attendance repository.AttendanceRepository,
	sessions repository.SessionRepository,
	students repository.StudentRepository,
	classes repository.ClassRepository,
	schools repository.SchoolRepository,
) *AttendanceUsecase {
	return &AttendanceUsecase{
		attendance: attendance,
		sessions:   sessions,
		students:   students,
		classes:    classes,
		schools:    schools,
	}
}

func (u *AttendanceUsecase) weeklySessionForOwner(ctx context.Context, actor params.OwnerActor, weeklySessionID types.ID) (entity.Session, messages.Message) {
	if !weeklySessionID.Valid() {
		return entity.Session{}, message.InvalidSessionInput
	}
	if !actor.Valid() {
		return entity.Session{}, message.Forbidden
	}
	sess, err := u.sessions.ReadByID(ctx, weeklySessionID)
	if err != nil {
		log.Printf("attendance weekly session read: %v", err)
		return entity.Session{}, messages.SharedServerError()
	}
	if !sess.ID.Valid() {
		return entity.Session{}, message.SessionNotFound
	}
	school, err := u.schools.ReadByIDForOwner(ctx, sess.SchoolID, actor.UserID)
	if err != nil {
		log.Printf("attendance owner school: %v", err)
		return entity.Session{}, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return entity.Session{}, message.Forbidden
	}
	return sess, messages.OK()
}

func (u *AttendanceUsecase) ensureRoll(ctx context.Context, sess entity.Session) messages.Message {
	_, err := u.attendance.EnsureForWeeklySession(ctx, sess.ID, sess.ClassID, sess.SessionDate)
	if err != nil {
		log.Printf("attendance ensure roll: %v", err)
		return messages.SharedServerError()
	}
	return messages.OK()
}

func (u *AttendanceUsecase) Read(ctx context.Context, p params.ReadAttendanceParams) (entity.AttendanceRoll, messages.Message) {
	sess, msg := u.weeklySessionForOwner(ctx, p.Actor, p.WeeklySessionID)
	if msg.Error != nil {
		return entity.AttendanceRoll{}, msg
	}
	records, err := u.attendance.ListRecords(ctx, sess.ID)
	if err != nil {
		log.Printf("attendance read list: %v", err)
		return entity.AttendanceRoll{}, messages.SharedServerError()
	}
	return entity.AttendanceRoll{
		SessionID: sess.ID,
		ClassID:   sess.ClassID,
		Records:   records,
	}, message.AttendanceRead
}

func (u *AttendanceUsecase) SetRecord(ctx context.Context, p params.SetAttendanceRecordParams) (entity.AttendanceRecord, messages.Message) {
	sess, msg := u.weeklySessionForOwner(ctx, p.Actor, p.WeeklySessionID)
	if msg.Error != nil {
		return entity.AttendanceRecord{}, msg
	}
	status := validation.NormalizeAttendanceStatus(p.Status)
	if !validation.ValidAttendanceStatus(status) {
		return entity.AttendanceRecord{}, message.InvalidAttendanceInput
	}
	if !p.StudentID.Valid() {
		return entity.AttendanceRecord{}, message.InvalidAttendanceInput
	}
	student, err := u.students.ReadByID(ctx, p.StudentID)
	if err != nil {
		log.Printf("attendance student read: %v", err)
		return entity.AttendanceRecord{}, messages.SharedServerError()
	}
	if !student.ID.Valid() || student.ClassID != sess.ClassID {
		return entity.AttendanceRecord{}, message.StudentNotFound
	}
	if msg = u.ensureRoll(ctx, sess); msg.Error != nil {
		return entity.AttendanceRecord{}, msg
	}
	if err := u.attendance.UpsertRecord(ctx, sess.ID, student.ID, status); err != nil {
		log.Printf("attendance set record: %v", err)
		return entity.AttendanceRecord{}, messages.SharedServerError()
	}
	return entity.AttendanceRecord{StudentID: student.ID, Status: status}, message.AttendanceSaved
}
