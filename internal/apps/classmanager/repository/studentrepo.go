package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/entity"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/database"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type StudentRepository interface {
	Create(ctx context.Context, classID types.ID, firstName, lastName, nationalCode, fatherName, studentNumber string) (entity.Student, error)
	ListForClass(ctx context.Context, classID types.ID) ([]entity.Student, error)
	ReadByID(ctx context.Context, studentID types.ID) (entity.Student, error)
	Update(ctx context.Context, studentID, classID types.ID, firstName, lastName, nationalCode, fatherName, studentNumber string) (entity.Student, error)
	Delete(ctx context.Context, studentID, classID types.ID) error
	NationalCodeTaken(ctx context.Context, nationalCode string, excludeStudentID types.ID) (bool, error)
}

type studentRepo struct {
	db *sql.DB
}

var _ StudentRepository = (*studentRepo)(nil)

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepo{db: db}
}

func nationalCodeDB(code string) any {
	if code == "" {
		return nil
	}
	return code
}

func (r *studentRepo) Create(ctx context.Context, classID types.ID, firstName, lastName, nationalCode, fatherName, studentNumber string) (entity.Student, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO students (class_id, first_name, last_name, national_code, father_name, student_number, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		classID, firstName, lastName, nationalCodeDB(nationalCode), fatherName, studentNumber, now, now)
	if err != nil {
		log.Printf("student create: %v", err)
		return entity.Student{}, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return entity.Student{}, err
	}
	return entity.Student{
		ID:            types.ID(id64),
		ClassID:       classID,
		FirstName:     firstName,
		LastName:      lastName,
		NationalCode:  nationalCode,
		FatherName:    fatherName,
		StudentNumber: studentNumber,
	}, nil
}

func (r *studentRepo) ListForClass(ctx context.Context, classID types.ID) ([]entity.Student, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, class_id, first_name, last_name, national_code, father_name, student_number
		FROM students
		WHERE class_id = ?
		ORDER BY last_name ASC, first_name ASC, id ASC`,
		classID)
	if err != nil {
		log.Printf("student list: %v", err)
		return nil, err
	}
	defer rows.Close()
	var out []entity.Student
	for rows.Next() {
		s, err := scanStudentRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *studentRepo) ReadByID(ctx context.Context, studentID types.ID) (entity.Student, error) {
	return r.scanStudent(r.db.QueryRowContext(ctx, `
		SELECT id, class_id, first_name, last_name, national_code, father_name, student_number
		FROM students WHERE id = ?`, studentID))
}

func (r *studentRepo) Update(ctx context.Context, studentID, classID types.ID, firstName, lastName, nationalCode, fatherName, studentNumber string) (entity.Student, error) {
	now := formatSQLiteTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, `
		UPDATE students
		SET first_name = ?, last_name = ?, national_code = ?, father_name = ?, student_number = ?, updated_at = ?
		WHERE id = ? AND class_id = ?`,
		firstName, lastName, nationalCodeDB(nationalCode), fatherName, studentNumber, now, studentID, classID)
	if err != nil {
		log.Printf("student update: %v", err)
		return entity.Student{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return entity.Student{}, sql.ErrNoRows
	}
	updated, err := r.ReadByID(ctx, studentID)
	if err != nil {
		return entity.Student{
			ID:            studentID,
			ClassID:       classID,
			FirstName:     firstName,
			LastName:      lastName,
			NationalCode:  nationalCode,
			FatherName:    fatherName,
			StudentNumber: studentNumber,
		}, nil
	}
	return updated, nil
}

func (r *studentRepo) Delete(ctx context.Context, studentID, classID types.ID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM students WHERE id = ? AND class_id = ?`, studentID, classID)
	if err != nil {
		log.Printf("student delete: %v", err)
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *studentRepo) NationalCodeTaken(ctx context.Context, nationalCode string, excludeStudentID types.ID) (bool, error) {
	if nationalCode == "" {
		return false, nil
	}
	row := r.db.QueryRowContext(ctx, `
		SELECT id FROM students WHERE national_code = ? AND id != ? LIMIT 1`,
		nationalCode, excludeStudentID)
	var id types.ID
	err := row.Scan(&id)
	if database.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return id.Valid(), nil
}

func scanStudentRow(sc interface {
	Scan(dest ...any) error
}) (entity.Student, error) {
	var s entity.Student
	var national sql.NullString
	err := sc.Scan(&s.ID, &s.ClassID, &s.FirstName, &s.LastName, &national, &s.FatherName, &s.StudentNumber)
	if err != nil {
		return entity.Student{}, err
	}
	if national.Valid {
		s.NationalCode = national.String
	}
	return s, nil
}

func (r *studentRepo) scanStudent(row *sql.Row) (entity.Student, error) {
	s, err := scanStudentRow(row)
	if database.IsNoRows(err) {
		return entity.Student{}, nil
	}
	if err != nil {
		return entity.Student{}, err
	}
	return s, nil
}
