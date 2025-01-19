package models

// structs for the batch

type JSONBatStructToDesReq struct {
	CorrelationID int64  `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type JSONBatStructToSerResp struct {
	CorrelationID int64  `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type JSONBatStructIdOrSh struct {
	CorrelationID int64  `json:"-"`
	OriginalURL   string `json:"-"`
	ShortURL      string `json:"-"`
}
