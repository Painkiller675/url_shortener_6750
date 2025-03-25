// Package contains DTOs.
package models

import "github.com/golang-jwt/jwt/v4"

// structs for the batch
// JSONBatStructToDesReq is used to deserialize request in a batch endpoint.
type JSONBatStructToDesReq struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// JSONBatStructToSerResp is used to serialize data in a batch endpoint.
type JSONBatStructToSerResp struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// JSONBatStructIDOrSh is used for data transformation.
type JSONBatStructIDOrSh struct {
	CorrelationID string `json:"-"`
	OriginalURL   string `json:"-"`
	ShortURL      string `json:"-"`
}

// UserURLS is used for short URLs and original URLs.
type UserURLS struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLIsDel is used to mark URL as deleted.
type URLIsDel struct {
	URL   string `json:"-"`
	IsDel bool   `json:"-"`
}

// Claims — структура утверждений, которая включает стандартные утверждения и одно пользовательское UserID.
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
