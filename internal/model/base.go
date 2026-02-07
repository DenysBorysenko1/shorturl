package model

type BaseEntity struct {
	ID string
}

func (entity BaseEntity) GetID() string {
	return entity.ID
}
