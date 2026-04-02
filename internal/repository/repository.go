package repository

type Entity interface {
	GetID() string
}

type Repository[T Entity] interface {
	Create(entity T) error
	CreateMany(entity []T) (int, error)
	GetByID(id string) (T, error)
	GetByURL(url string) (T, error)
	GetAllByUserID(userID string) ([]T, error)
	SoftDeleteByIDs(ids []string, userID string) error
}
