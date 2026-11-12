package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type DisciplineRepository interface {
	Create(ctx context.Context, classID, studentID types.ID, studentName, recordType, note, recordDate string) (entity.DisciplineRecord, error)
	ListForSchool(ctx context.Context, schoolID types.ID, classID, studentID types.ID) ([]entity.DisciplineRecord, error)
	ReadByID(ctx context.Context, recordID types.ID) (entity.DisciplineRecord, error)
	Update(ctx context.Context, recordID, classID, studentID types.ID, studentName, recordType, note, recordDate string) (entity.DisciplineRecord, error)
	Delete(ctx context.Context, recordID types.ID) error
}

type disciplineRepo struct {
	db *sql.DB
}

var _ DisciplineRepository = (*disciplineRepo)(nil)

func NewDisciplineRepository(db *sql.DB) DisciplineRepository {
	return &disciplineRepo{db: db}
}

const disciplineSelect = `
	SELECT dr.id, dr.class_id, dr.student_id, dr.student_name, dr.type, dr.note, dr.record_date
	FROM discipline_records dr
	INNER JOIN classes c ON c.id = dr.class_id`

func scanDiscipline(sc interface {
	Scan(dest ...any) error
}) (entity.DisciplineRecord, error) {
	var d entity.DisciplineRecord
	err := sc.Scan(&d.ID, &d.ClassID, &d.StudentID, &d.StudentName, &d.Type, &d.Note, &d.RecordDate)
	return d, err
}

func scanDisciplineRows(rows *sql.Rows) ([]entity.DisciplineRecord, error) {
	var out []entity.DisciplineRecord
	for rows.Next() {
		d, err := scanDiscipline(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *disciplineRepo) Create(ctx context.Context, classID, studentID types.ID, studentName, recordType, note, recordDate string) (entity.DisciplineRecord, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO discipline_records (class_id, student_id, student_name, type, note, record_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		classID, studentID, studentName, recordType, note, recordDate, now, now)
	if err != nil {
		log.Printf("discipline create: %v", err)
		return entity.DisciplineRecord{}, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return entity.DisciplineRecord{}, err
	}
	return entity.DisciplineRecord{
		ID:          types.ID(id64),
		ClassID:     classID,
		StudentID:   studentID,
		StudentName: studentName,
		Type:        recordType,
		Note:        note,
		RecordDate:  recordDate,
	}, nil
}

func (r *disciplineRepo) ListForSchool(ctx context.Context, schoolID types.ID, classID, studentID types.ID) ([]entity.DisciplineRecord, error) {
	q := disciplineSelect + `
		WHERE c.school_id = ?`
	args := []any{schoolID}
	if classID.Valid() {
		q += ` AND dr.class_id = ?`
		args = append(args, classID)
	}
	if studentID.Valid() {
		q += ` AND dr.student_id = ?`
		args = append(args, studentID)
	}
	q += ` ORDER BY dr.record_date DESC, dr.id DESC`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		log.Printf("discipline list: %v", err)
		return nil, err
	}
	defer rows.Close()
	return scanDisciplineRows(rows)
}

func (r *disciplineRepo) ReadByID(ctx context.Context, recordID types.ID) (entity.DisciplineRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, class_id, student_id, student_name, type, note, record_date
		FROM discipline_records
		WHERE id = ?`, recordID)
	d, err := scanDiscipline(row)
	if err == sql.ErrNoRows {
		return entity.DisciplineRecord{}, sql.ErrNoRows
	}
	if err != nil {
		log.Printf("discipline read: %v", err)
		return entity.DisciplineRecord{}, err
	}
	return d, nil
}

func (r *disciplineRepo) Update(ctx context.Context, recordID, classID, studentID types.ID, studentName, recordType, note, recordDate string) (entity.DisciplineRecord, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		UPDATE discipline_records
		SET class_id = ?, student_id = ?, student_name = ?, type = ?, note = ?, record_date = ?, updated_at = ?
		WHERE id = ?`,
		classID, studentID, studentName, recordType, note, recordDate, now, recordID)
	if err != nil {
		log.Printf("discipline update: %v", err)
		return entity.DisciplineRecord{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return entity.DisciplineRecord{}, sql.ErrNoRows
	}
	return entity.DisciplineRecord{
		ID:          recordID,
		ClassID:     classID,
		StudentID:   studentID,
		StudentName: studentName,
		Type:        recordType,
		Note:        note,
		RecordDate:  recordDate,
	}, nil
}

func (r *disciplineRepo) Delete(ctx context.Context, recordID types.ID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM discipline_records WHERE id = ?`, recordID)
	if err != nil {
		log.Printf("discipline delete: %v", err)
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
