package order

type Repository interface {
	GetByID(id int64) (Order, bool)
}

type MemoryRepository struct {
	orders map[int64]Order
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		orders: map[int64]Order{
			101: {ID: 101, UserID: 1, Item: "Ноутбук", Price: 79990},
			102: {ID: 102, UserID: 2, Item: "Мышь", Price: 2490},
			103: {ID: 103, UserID: 1, Item: "Клавиатура", Price: 5990},
		},
	}
}

func (r *MemoryRepository) GetByID(id int64) (Order, bool) {
	order, ok := r.orders[id]
	return order, ok
}
