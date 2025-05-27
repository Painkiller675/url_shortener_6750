package controller

import (
	"context"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
)

type Business interface {
	PingDB(context.Context) error
	StoreAlURL(ctx context.Context, alias string, url string, userID string) (int64, error)
	GetOrURLByAl(ctx context.Context, alias string) (string, error)
	SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error)
}
