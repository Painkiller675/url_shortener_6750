// the package is for getting the access to the business logic via Business interface
package business

import (
	"context"
	"github.com/Painkiller675/url_shortener_6750/internal/models"

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

// StoreAlURL is used for / handler
func (b *Business) StoreAlURL(ctx context.Context, alias string, url string, userID string) (int64, error) {
	return b.Storage.StoreAlURL(ctx, alias, url, userID)
}

// GetOrURLByAl is used for /{id} handler
func (b *Business) GetOrURLByAl(ctx context.Context, alias string) (string, error) {
	return b.Storage.GetOrURLByAl(ctx, alias)
}

// SaveBatchURL is used in /api/shorten/batch handler
func (b *Business) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	return b.Storage.SaveBatchURL(ctx, corURLSh)
}
