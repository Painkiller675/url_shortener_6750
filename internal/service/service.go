// Package is used for generation random strings (8 letters)
package service

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/Painkiller675/url_shortener_6750/internal/models"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GetRandString generates sha1 hash and cut it to 8 letters.
func GetRandString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	sha1Hash := hex.EncodeToString(h.Sum(nil))
	sha1Hash8 := sha1Hash[:8]
	return sha1Hash8

}

// CreateBatchIDOrSh transform data from API into inner-used type.
func CreateBatchIDOrSh(desBatchReq *[]models.JSONBatStructToDesReq) (*[]models.JSONBatStructIDOrSh, error) {
	// allocate memory for an auxiliary array of structs
	idURLSh := make([]models.JSONBatStructIDOrSh, 0) // TODO [MENTOR]: is it well-allocated? Why len(desBatchReq) instead of 0 is false??
	// filling the array
	for _, idURL := range *desBatchReq {
		randAl := GetRandString(idURL.OriginalURL)
		idURLSh = append(idURLSh, models.JSONBatStructIDOrSh{
			CorrelationID: idURL.CorrelationID,
			OriginalURL:   idURL.OriginalURL,
			ShortURL:      randAl,
		})

	}

	// returning the batch for response
	return &idURLSh, nil
}
