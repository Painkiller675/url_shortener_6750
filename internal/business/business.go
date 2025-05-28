// the package is for getting the access to the business logic via Business interface
package business

import (
	"context"
	"errors"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"github.com/Painkiller675/url_shortener_6750/internal/service"
	"net/url"

	"github.com/Painkiller675/url_shortener_6750/internal/repository"
)

// Business is used to implement all the methods with the business logic
type Business struct {
	Storage repository.URLStorage
}

// PingDB is used to figure out if the database is available
func (b *Business) PingDB(ctx context.Context) error {
	return b.Storage.Ping(ctx)
}

// StoreAlURL is used for / handler it returns molded URL and an error if any
func (b *Business) StoreAlURL(ctx context.Context, urlIn *url.URL, userID string) (string, error) {
	//errs := make([]error, 0)
	randAl := service.GetRandString(urlIn.String()) // TODO: replace it to business
	_, err := b.Storage.StoreAlURL(ctx, randAl, urlIn.String(), userID)
	//errs = append(errs, err)
	baseURL := config.StartOptions.BaseURL
	baseURL.JoinPath(randAl) // resultURL
	//resultURL, err := url.JoinPath(baseURL, randAl) // err == nil anyway because we have already parsed the url in the  handler
	//errs = append(errs, err)
	return baseURL, errors.Join(errs...) // I return resultURL anyway
}

// GetOrURLByAl is used for /{id} handler
func (b *Business) GetOrURLByAl(ctx context.Context, alias string) (string, error) {
	return b.Storage.GetOrURLByAl(ctx, alias)
}

// SaveBatchURL is used in /api/shorten/batch handler
func (b *Business) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	return b.Storage.SaveBatchURL(ctx, corURLSh)
}
