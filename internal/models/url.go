package models

// structs for the batch

type JSONBatStructToDesReq struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type JSONBatStructToSerResp struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type JSONBatStructIDOrSh struct {
	CorrelationID string `json:"-"`
	OriginalURL   string `json:"-"`
	ShortURL      string `json:"-"`
}

/*type ExistsURLError struct {
	ExistedAlias string
	Err          error
}

func (e *ExistsURLError) Error() string {
	return fmt.Sprintf("[%s:%v]", e.ExistedAlias, e.Err)
}
*/
