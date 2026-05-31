package user

type Repository interface {
	GetByID(id int64) (User, bool)
}

type MemoryRepository struct {
	users map[int64]User
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users: map[int64]User{
			1: {ID: 1, Name: "Иван Иванов", Email: "ivan@example.com"},
			2: {ID: 2, Name: "Мария Петрова", Email: "maria@example.com"},
			3: {ID: 3, Name: "Алексей Сидоров", Email: "alex@example.com"},
		},
	}
}

func (r *MemoryRepository) GetByID(id int64) (User, bool) {
	user, ok := r.users[id]
	return user, ok
}
