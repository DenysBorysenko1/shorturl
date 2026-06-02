package model

type BaseEntity struct {
	ID string `db:"id"`
}

func (entity BaseEntity) GetID() string {
	return entity.ID
}
