package repository

type Entity interface {
	GetID() string
}

type Repository[T Entity] interface {
	Create(entity T) error
	CreateMany(entity []T) (int, error)
	GetByID(id string) (T, error)
}
