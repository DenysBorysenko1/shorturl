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
