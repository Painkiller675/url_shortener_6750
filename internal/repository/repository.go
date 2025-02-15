package repository

import (
	"context"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"github.com/Painkiller675/url_shortener_6750/internal/repository/file"
	"github.com/Painkiller675/url_shortener_6750/internal/repository/memory"
	"github.com/Painkiller675/url_shortener_6750/internal/repository/pg"
	_ "github.com/jackc/pgx"
	"go.uber.org/zap"
)

type URLStorage interface {
	//StoreAlURL(ctx context.Context, alias string, url string) error
	StoreAlURL(ctx context.Context, alias string, url string, userID string) (int64, error)
	GetOrURLByAl(ctx context.Context, alias string) (string, error)
	Ping(ctx context.Context) error
	SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error)
	GetAlByURL(ctx context.Context, url string) (string, error)
	GetDataByUserID(ctx context.Context, userID string) (*[]models.UserURLS, error)
	DeleteURLsByUserID(ctx context.Context, userID string, aliasToDel []string) error
	CheckIfUserExists(ctx context.Context, userID string) error
}

func ChooseStorage(ctx context.Context, logger *zap.Logger) (URLStorage, error) {
	// if the database storage
	if config.StartOptions.DBConStr != "" {

		pgStor, err := pg.NewStorage(ctx, config.StartOptions.DBConStr, logger)
		if err != nil {
			logger.Error("[ERROR] Can't open pg database ", zap.Error(err))
			return nil, err // TODO: [4 MENTOR] unuseful cause I use only panic in constructor in fact, is it ok?
		}

		return pgStor, nil
	}
	// if the file storage
	if config.StartOptions.Filename != "" {
		return file.NewStorage(config.StartOptions.Filename, logger), nil
	}
	// if the memory storage
	return memory.NewStorage(logger), nil

}
