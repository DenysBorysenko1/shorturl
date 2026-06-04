// Package model provides data structures and entities
// for the application, including base entities and domain models.
package model

// BaseEntity provides a common ID field for all entities.
//
// This is embedded in other model types to provide automatic
// unique identifier support.
type BaseEntity struct {
	ID string `db:"id"` // Unique identifier for the entity
}

// GetID returns the unique identifier of the entity.
func (entity BaseEntity) GetID() string {
	return entity.ID
}
