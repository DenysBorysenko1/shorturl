package dto

// RequestBatchItem represents an item in a batch shortening request.
//
// Each item contains a correlation ID that will be preserved in the response
// along with the original URL to be shortened.
type RequestBatchItem struct {
	CorrelationID string `json:"correlation_id"` // Client-provided identifier for this URL
	OriginalURL   string `json:"original_url"`   // The original URL to shorten
}

// ResponseBatchItem represents an item in a batch shortening response.
//
// Each item contains the correlation ID from the request and the generated
// short URL.
type ResponseBatchItem struct {
	CorrelationID string `json:"correlation_id"` // The correlation ID from the request
	ShortURL      string `json:"short_url"`      // The generated short URL
}
