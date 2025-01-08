package main

import (
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/controller"
	gzipMW "github.com/Painkiller675/url_shortener_6750/internal/middleware/gzip"
	"github.com/Painkiller675/url_shortener_6750/internal/middleware/logger"
	"github.com/Painkiller675/url_shortener_6750/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	// init config
	config.SetConfig()

	// init logger
	l, err := logger.NewZapLogger(config.StartOptions.LogLvl)
	if err != nil {
		log.Panic(err)
	}

	// init storage
	s := repository.NewStorage(l.Logger)

	// init controller
	c := controller.New(l.Logger, s)

	// init router
	r := chi.NewRouter()

	// set logger for chi router
	r.Use(l.LogMW)
	r.Use(gzipMW.GzMW)

	// routing
	r.Route("/", func(r chi.Router) {
		r.Post("/", c.CreateShortURLHandler)
		r.Get("/{id}", c.GetLongURLHandler)
		r.Post("/api/shorten", c.CreateShortURLJSONHandler)
	})
	//start server
	l.Logger.Info("Running server", zap.String("address", config.StartOptions.HTTPServer.Address))
	if err := http.ListenAndServe(config.StartOptions.HTTPServer.Address, r); err != nil {
		panic(err)
	}

}
