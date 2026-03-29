package model

type Link struct {
	BaseEntity
	URL       string `db:"url"`
	CreatedBy string `db:"created_by"`
}

func (entity Link) GetURL() string {
	return entity.URL
}

func (entity Link) GetCreatedBy() string {
	return entity.CreatedBy
}
