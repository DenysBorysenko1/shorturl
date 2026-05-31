package repository

// Entity defines the interface for entities that can be stored in a repository.
//
// Any entity stored in a repository must implement this interface to provide
// a unique identifier.
type Entity interface {
	// GetID returns the unique identifier for this entity.
	GetID() string
}

// Repository defines the interface for generic data storage and retrieval operations.
//
// This is a generic repository interface that can work with any type that implements
// the Entity interface. It provides CRUD operations and user-specific queries.
type Repository[T Entity] interface {
	// Create stores a single entity in the repository.
	Create(entity T) error
	// CreateMany stores multiple entities in a single batch operation.
	// Returns the number of entities successfully created.
	CreateMany(entity []T) (int, error)
	// GetByID retrieves an entity by its unique identifier.
	GetByID(id string) (T, error)
	// GetByURL retrieves an entity by its URL field.
	GetByURL(url string) (T, error)
	// GetAllByUserID retrieves all entities created by a specific user.
	GetAllByUserID(userID string) ([]T, error)
	// SoftDeleteByIDs marks multiple entities as deleted without removing them from storage.
	// Only entities belonging to the specified user will be affected.
	SoftDeleteByIDs(ids []string, userID string) error
}
