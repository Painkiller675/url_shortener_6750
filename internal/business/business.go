// the package is for getting the access to the business logic via Business interface
package business

import (
	"context"
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

// CheckIfUserExists checks if a user exists in the database. It's used in DeleteURLSHandler handler (GET /api/user/urls)
func (b *Business) CheckIfUserExists(ctx context.Context, userID string) error {
	return b.Storage.CheckIfUserExists(ctx, userID)
}

// GetStats is used to get the statistics namely the number of urls and users in the database - /api/internal/stats
func (b *Business) GetStats(ctx context.Context) (urls int, users int, err error) {
	return b.Storage.GetStats(ctx)
}

// PingDB is used to figure out if the database is available
func (b *Business) PingDB(ctx context.Context) error {
	return b.Storage.Ping(ctx)
}

// StoreAlURL is used for / handler it returns molded URL and an error if any
func (b *Business) StoreAlURL(ctx context.Context, urlIn *url.URL, userID string) (string, error) {
	//errs := make([]error, 0)
	urlInStr := urlIn.String()                // TODO: привести к стрингу тут или слать приведение везде?
	randAl := service.GetRandString(urlInStr) // TODO: mb replace it to business
	_, err := b.Storage.StoreAlURL(ctx, randAl, urlInStr, userID)
	//errs = append(errs, err)
	baseURL := config.StartOptions.BaseURL
	baseURL = baseURL.JoinPath(randAl) // resultURL // TODO: reassign??
	//resultURL, err := url.JoinPath(baseURL, randAl) // err == nil anyway because we have already parsed the url in the  handler
	//errs = append(errs, err)
	return baseURL.String(), err // I return resultURL anyway
}

// GetOrURLByAl is used for /{id}  (in REST controller - GetLongURLHandler handler)
func (b *Business) GetOrURLByAl(ctx context.Context, alias string) (string, error) {
	return b.Storage.GetOrURLByAl(ctx, alias)
}

// SaveBatchURL is used in /api/shorten/batch (REST CreateShortURLJSONBatchHandler handler)
func (b *Business) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	return b.Storage.SaveBatchURL(ctx, corURLSh)
}

// GetDataByUserID  /api/user/urls returns aliases of a particular user (GetUserURLSHandler handler)
func (b *Business) GetDataByUserID(ctx context.Context, userID string) (*[]models.UserURLS, error) {
	return b.Storage.GetDataByUserID(ctx, userID)
}
