package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/controller"
	gzipMW "github.com/Painkiller675/url_shortener_6750/internal/middleware/gzip"
	"github.com/Painkiller675/url_shortener_6750/internal/middleware/logger"
	"github.com/Painkiller675/url_shortener_6750/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	// init config
	config.SetConfig()

	// init logger
	l, err := logger.NewZapLogger(config.StartOptions.LogLvl)
	if err != nil {
		log.Panic(err)
	}
	l.Logger.Info("Starting server", zap.String("ConString: ", config.StartOptions.DBConStr), zap.String("BaseURL:", config.StartOptions.BaseURL))
	//render logger for gzip
	//gzipMW.NewGzipLogger(l.Logger)

	//init the context
	ctx := context.Background()

	// init storage
	s, err := repository.ChooseStorage(ctx, l.Logger)
	if err != nil {
		panic(err) // TODO: [MENTOR] is it good to panic here or I could handle it miles better?
	}

	chanJobs := make(chan controller.JobToDelete, 100)
	defer close(chanJobs)

	go deleteUrl(s, chanJobs)
	// create a wait group
	var wg sync.WaitGroup // TODO bring it to controller
	// init controller
	c := controller.New(l.Logger, s, chanJobs) //

	// init router
	r := chi.NewRouter()

	// set logger for chi router
	r.Use(l.LogMW)
	r.Use(gzipMW.GzMW)

	// routing
	r.Route("/", func(r chi.Router) {
		r.Post("/", c.CreateShortURLHandler())
		r.Get("/ping", c.PingDB())
		r.Get("/{id}", c.GetLongURLHandler())
		r.Post("/api/shorten", c.CreateShortURLJSONHandler())
		r.Post("/api/shorten/batch", c.CreateShortURLJSONBatchHandler())
		r.Get("/api/user/urls", c.GetUserURLSHandler())
		r.Delete("/api/user/urls", c.DeleteURLSHandler())

	})
	//start server
	l.Logger.Info("Running server", zap.String("address", config.StartOptions.HTTPServer.Address))
	if err := http.ListenAndServe(config.StartOptions.HTTPServer.Address, r); err != nil {
		panic(err)
	}
	wg.Wait() // gracefull shutdown
}

func deleteUrl(s repository.URLStorage, jobs chan controller.JobToDelete) {
	for job := range jobs {
		if err := s.DeleteURLsByUserID(context.Background(), job.UserId, job.LsUrl); err != nil {
			fmt.Println("[ERROR]", zap.Error(err)) // TODO [MENTOR]: how to go it up? is it necessary?
		}
	}
}
