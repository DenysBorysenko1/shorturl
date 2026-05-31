package dto

// ResponseListItem represents a single URL entry in the user's URL list response.
//
// This structure contains both the original URL and its corresponding short URL,
// returned when requesting all URLs created by a user.
type ResponseListItem struct {
	ShortURL    string `json:"short_url"`    // The short URL
	OriginalURL string `json:"original_url"` // The original URL that was shortened
}
