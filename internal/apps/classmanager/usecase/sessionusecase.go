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

type SessionUsecase struct {
	sessions repository.SessionRepository
	classes  repository.ClassRepository
	schools  repository.SchoolRepository
}

func NewSessionUsecase(sessions repository.SessionRepository, classes repository.ClassRepository, schools repository.SchoolRepository) *SessionUsecase {
	return &SessionUsecase{sessions: sessions, classes: classes, schools: schools}
}

func (u *SessionUsecase) classForOwner(ctx context.Context, actor params.OwnerActor, classID types.ID) (entity.Class, messages.Message) {
	if !classID.Valid() {
		return entity.Class{}, message.InvalidClassInput
	}
	if !actor.Valid() {
		return entity.Class{}, message.Forbidden
	}
	class, err := u.classes.ReadByID(ctx, classID)
	if err != nil {
		log.Printf("session class read: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	if !class.ID.Valid() {
		return entity.Class{}, message.ClassNotFound
	}
	school, err := u.schools.ReadByIDForOwner(ctx, class.SchoolID, actor.UserID)
	if err != nil {
		log.Printf("session owner school check: %v", err)
		return entity.Class{}, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return entity.Class{}, message.Forbidden
	}
	return class, messages.OK()
}

func (u *SessionUsecase) resolveSchoolID(ctx context.Context, actor params.OwnerActor, schoolID types.ID) (types.ID, messages.Message) {
	if !actor.Valid() {
		return 0, message.Forbidden
	}
	if schoolID.Valid() {
		school, err := u.schools.ReadByIDForOwner(ctx, schoolID, actor.UserID)
		if err != nil {
			log.Printf("session school owner check: %v", err)
			return 0, messages.SharedServerError()
		}
		if !school.ID.Valid() {
			return 0, message.Forbidden
		}
		return school.ID, messages.OK()
	}
	school, err := u.schools.ReadFirstForOwner(ctx, actor.UserID)
	if err != nil {
		log.Printf("session resolve first school: %v", err)
		return 0, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return 0, message.NoSchool
	}
	return school.ID, messages.OK()
}

func (u *SessionUsecase) sessionForOwner(ctx context.Context, actor params.OwnerActor, sessionID types.ID) (entity.Session, messages.Message) {
	if !sessionID.Valid() {
		return entity.Session{}, message.InvalidSessionInput
	}
	if !actor.Valid() {
		return entity.Session{}, message.Forbidden
	}
	session, err := u.sessions.ReadByID(ctx, sessionID)
	if err != nil {
		log.Printf("session read: %v", err)
		return entity.Session{}, messages.SharedServerError()
	}
	if !session.ID.Valid() {
		return entity.Session{}, message.SessionNotFound
	}
	school, err := u.schools.ReadByIDForOwner(ctx, session.SchoolID, actor.UserID)
	if err != nil {
		log.Printf("session owner check: %v", err)
		return entity.Session{}, messages.SharedServerError()
	}
	if !school.ID.Valid() {
		return entity.Session{}, message.Forbidden
	}
	return session, messages.OK()
}

type normalizedSessionFields struct {
	sessionDate string
	day         int
	start       string
	end         string
	subject     string
}

func (u *SessionUsecase) normalizeFields(sessionDate string, day int, start, end, subject string) (normalizedSessionFields, messages.Message) {
	sessionDate = validation.NormalizeSessionDate(sessionDate)
	start = validation.NormalizeClock(start)
	end = validation.NormalizeClock(end)
	subject = validation.NormalizeSessionSubject(subject)

	if !validation.ValidSessionDate(sessionDate) ||
		!validation.ValidSessionTimeRange(start, end) ||
		!validation.ValidSessionSubject(subject) {
		return normalizedSessionFields{}, message.InvalidSessionInput
	}
	if !validation.ValidSessionDay(day) {
		return normalizedSessionFields{}, message.InvalidSessionInput
	}
	if !validation.DayMatchesSessionDate(day, sessionDate) {
		return normalizedSessionFields{}, message.SessionDayMismatch
	}
	return normalizedSessionFields{
		sessionDate: sessionDate,
		day:         day,
		start:       start,
		end:         end,
		subject:     subject,
	}, messages.OK()
}

func (u *SessionUsecase) ensureNoOverlap(ctx context.Context, schoolID types.ID, sessionDate, start, end string, exclude types.ID) messages.Message {
	overlap, err := u.sessions.HasTimeOverlap(ctx, schoolID, sessionDate, start, end, exclude)
	if err != nil {
		log.Printf("session overlap check: %v", err)
		return messages.SharedServerError()
	}
	if overlap {
		return message.SessionTimeOverlap
	}
	return messages.OK()
}

func (u *SessionUsecase) List(ctx context.Context, p params.ListSessionsParams) ([]entity.Session, messages.Message) {
	if p.ClassID.Valid() {
		if _, msg := u.classForOwner(ctx, p.Actor, p.ClassID); msg.Error != nil {
			return nil, msg
		}
		list, err := u.sessions.ListForClass(ctx, p.ClassID)
		if err != nil {
			log.Printf("session list usecase: %v", err)
			return nil, messages.SharedServerError()
		}
		return list, message.SessionRead
	}
	schoolID, msg := u.resolveSchoolID(ctx, p.Actor, p.SchoolID)
	if msg.Error != nil {
		return nil, msg
	}
	list, err := u.sessions.ListForSchool(ctx, schoolID)
	if err != nil {
		log.Printf("session list usecase: %v", err)
		return nil, messages.SharedServerError()
	}
	return list, message.SessionRead
}

func (u *SessionUsecase) Read(ctx context.Context, p params.ReadSessionParams) (entity.Session, messages.Message) {
	return u.sessionForOwner(ctx, p.Actor, p.SessionID)
}

func (u *SessionUsecase) Create(ctx context.Context, p params.CreateSessionParams) (entity.Session, messages.Message) {
	class, msg := u.classForOwner(ctx, p.Actor, p.ClassID)
	if msg.Error != nil {
		return entity.Session{}, msg
	}
	schoolID, msg := u.resolveSchoolID(ctx, p.Actor, p.SchoolID)
	if msg.Error != nil {
		return entity.Session{}, msg
	}
	if class.SchoolID != schoolID {
		return entity.Session{}, message.Forbidden
	}
	fields, msg := u.normalizeFields(p.SessionDate, p.Day, p.Start, p.End, p.Subject)
	if msg.Error != nil {
		return entity.Session{}, msg
	}
	if msg = u.ensureNoOverlap(ctx, schoolID, fields.sessionDate, fields.start, fields.end, 0); msg.Error != nil {
		return entity.Session{}, msg
	}
	session, err := u.sessions.Create(ctx, schoolID, class.ID, fields.sessionDate, fields.day, fields.start, fields.end, fields.subject)
	if err != nil {
		if isSQLiteSessionConflict(err) {
			return entity.Session{}, message.SessionTimeOverlap
		}
		log.Printf("session create usecase: %v", err)
		return entity.Session{}, messages.SharedServerError()
	}
	return session, message.SessionCreated
}

func (u *SessionUsecase) Update(ctx context.Context, p params.UpdateSessionParams) (entity.Session, messages.Message) {
	existing, msg := u.sessionForOwner(ctx, p.Actor, p.SessionID)
	if msg.Error != nil {
		return entity.Session{}, msg
	}
	if _, msg := u.classForOwner(ctx, p.Actor, existing.ClassID); msg.Error != nil {
		return entity.Session{}, msg
	}
	fields, msg := u.normalizeFields(p.SessionDate, p.Day, p.Start, p.End, p.Subject)
	if msg.Error != nil {
		return entity.Session{}, msg
	}
	if msg = u.ensureNoOverlap(ctx, existing.SchoolID, fields.sessionDate, fields.start, fields.end, existing.ID); msg.Error != nil {
		return entity.Session{}, msg
	}
	updated, err := u.sessions.Update(ctx, existing.ID, existing.SchoolID, existing.ClassID, fields.sessionDate, fields.day, fields.start, fields.end, fields.subject)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.Session{}, message.SessionNotFound
		}
		if isSQLiteSessionConflict(err) {
			return entity.Session{}, message.SessionTimeOverlap
		}
		log.Printf("session update usecase: %v", err)
		return entity.Session{}, messages.SharedServerError()
	}
	return updated, message.SessionUpdated
}

func (u *SessionUsecase) Delete(ctx context.Context, p params.DeleteSessionParams) messages.Message {
	existing, msg := u.sessionForOwner(ctx, p.Actor, p.SessionID)
	if msg.Error != nil {
		return msg
	}
	err := u.sessions.Delete(ctx, existing.ID, existing.ClassID)
	if err != nil {
		if err == sql.ErrNoRows {
			return message.SessionNotFound
		}
		log.Printf("session delete usecase: %v", err)
		return messages.SharedServerError()
	}
	return message.SessionDeleted
}

func isSQLiteSessionConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") && strings.Contains(msg, "sessions")
}
