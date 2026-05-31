package student

import (
	"errors"

	"example.com/pz2-grpc/gen/studentpb"
)

var ErrStudentNotFound = errors.New("student not found")

type Repository struct {
	data map[int64]*studentpb.Student
}

func NewRepository() *Repository {
	return &Repository{
		data: map[int64]*studentpb.Student{
			1: {
				Id:       1,
				FullName: "Иванов Иван Иванович",
				Group:    "ИВБО-01-25",
				Email:    "ivanov@example.com",
			},
			2: {
				Id:       2,
				FullName: "Петрова Мария Сергеевна",
				Group:    "ИВБО-02-25",
				Email:    "petrova@example.com",
			},
			3: {
				Id:       3,
				FullName: "Сидоров Алексей Андреевич",
				Group:    "ИВБО-03-25",
				Email:    "sidorov@example.com",
			},
		},
	}
}

func (r *Repository) GetByID(id int64) (*studentpb.Student, error) {
	st, ok := r.data[id]
	if !ok {
		return nil, ErrStudentNotFound
	}

	return st, nil
}
