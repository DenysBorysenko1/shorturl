package repository

type Entity interface {
	GetID() string
}

type Repository[T Entity] interface {
	Create(entity T) error
	GetById(id string) (T, error)
}
