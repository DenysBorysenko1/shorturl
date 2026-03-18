package model

type Link struct {
	BaseEntity
	URL string
}

func (entity Link) GetURL() string {
	return entity.URL
}
