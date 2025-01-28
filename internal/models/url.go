package models

import "github.com/golang-jwt/jwt/v4"

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

type UserURLS struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

/*type ExistsURLError struct {
	ExistedAlias string
	Err          error
}

func (e *ExistsURLError) Error() string {
	return fmt.Sprintf("[%s:%v]", e.ExistedAlias, e.Err)
}
*/
