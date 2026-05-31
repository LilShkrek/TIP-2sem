package order

type Order struct {
	ID     int64   `json:"id"`
	UserID int64   `json:"user_id"`
	Item   string  `json:"item"`
	Price  float64 `json:"price"`
}

type UserDTO struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type OrderWithUser struct {
	Order Order   `json:"order"`
	User  UserDTO `json:"user"`
}
