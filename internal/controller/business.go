package controller

import (
	"context"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
)

type Business interface {
	PingDB(context.Context) error
	StoreAlURL(ctx context.Context, url string, userID string) (string, error)
	GetOrURLByAl(ctx context.Context, alias string) (string, error)
	SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error)
}
