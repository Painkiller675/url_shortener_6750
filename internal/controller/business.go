package controller

import (
	"context"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"net/url"
)

type Business interface {
	PingDB(context.Context) error
	StoreAlURL(ctx context.Context, url *url.URL, userID string) (string, error)
	GetOrURLByAl(ctx context.Context, alias string) (string, error)
	SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error)
	GetDataByUserID(ctx context.Context, userID string) (*[]models.UserURLS, error)
	CheckIfUserExists(ctx context.Context, userID string) error
	GetStats(ctx context.Context) (urls int, users int, err error)
}
