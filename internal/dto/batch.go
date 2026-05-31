package dto

type RequestBatchItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseBatchItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
