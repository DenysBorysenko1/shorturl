package repository

type EntityRepository[T any] interface {
	Create(entity T) (T, error)
	Get(id string) (T, error)
}

