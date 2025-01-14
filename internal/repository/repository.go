package repository

import (
	"context"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/repository/file"
	"github.com/Painkiller675/url_shortener_6750/internal/repository/memory"
	"github.com/Painkiller675/url_shortener_6750/internal/repository/pg"
	_ "github.com/jackc/pgx"
	"go.uber.org/zap"
)

type URLStorage interface {
	//StoreAlURL(ctx context.Context, alias string, url string) error
	StoreAlURL(ctx context.Context, alias string, url string) (int64, error)
	GetOrURLByAl(ctx context.Context, alias string) (string, error)
	Ping(ctx context.Context) error
}

func ChooseStorage(logger *zap.Logger) (URLStorage, error) {
	// if the database storage
	if config.StartOptions.DBConStr != "" {

		pgStor, err := pg.NewStorage(config.StartOptions.DBConStr)
		if err != nil {
			logger.Info("[ERROR] Can't open pg database ", zap.Error(err))
			return nil, err // TODO: [4 MENTOR] unuseful cause I use only panic in constructor in fact, is it ok?
		}
		err = pgStor.Bootstrap(context.Background()) //TODO: is it ok to use new contex here?
		if err != nil {
			logger.Info("[ERROR] Can't bootstrap pg database ", zap.Error(err))
			return nil, err
		}
		return pgStor, nil
	}
	// if the file storage
	if config.StartOptions.DBConStr != "" {
		return file.NewStorage(config.StartOptions.Filename, logger), nil
	}
	// if the memory storage
	return memory.NewStorage(logger), nil
	// TODO other storages
	// конфиг сюда и вернунуть urlstorage и error и вызывать буду в main тут я вызываю конструкторы файла или мемори и сразу возвращаю их в main
}
