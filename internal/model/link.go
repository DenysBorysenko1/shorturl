package model

// Link represents a shortened URL with its metadata.
//
// This structure contains all information about a short link including
// the original URL, the user who created it, and its deletion status.
type Link struct {
	BaseEntity
	URL       string `db:"url"`         // The original URL that was shortened
	CreatedBy string `db:"created_by"`   // ID of the user who created this link
	IsDeleted bool   `db:"is_deleted"`   // Whether the link has been soft-deleted
}

// GetURL returns the original URL of the link.
func (entity Link) GetURL() string {
	return entity.URL
}

// GetCreatedBy returns the ID of the user who created this link.
func (entity Link) GetCreatedBy() string {
	return entity.CreatedBy
}
