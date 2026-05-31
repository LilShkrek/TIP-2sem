package student

import (
	"database/sql"
	"errors"
)

var ErrStudentNotFound = errors.New("student not found")

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// UnsafeGetByID показывает опасный подход с конкатенацией SQL.
// Метод оставлен только для демонстрации SQL-инъекции и не используется в HTTP-обработчиках.
func (r *Repo) UnsafeGetByID(rawID string) (*Student, error) {
	query := "SELECT id, full_name, study_group, email FROM students WHERE id = " + rawID

	row := r.db.QueryRow(query)

	var st Student
	err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	return &st, nil
}

func (r *Repo) GetByID(id int64) (*Student, error) {
	row := r.db.QueryRow(
		"SELECT id, full_name, study_group, email FROM students WHERE id = $1",
		id,
	)

	var st Student
	err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	return &st, nil
}

func (r *Repo) PrepareGetByID() (*sql.Stmt, error) {
	return r.db.Prepare("SELECT id, full_name, study_group, email FROM students WHERE id = $1")
}
