package main

import (
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/handlers"
	gzipMW "github.com/Painkiller675/url_shortener_6750/internal/middleware/gzip"
	"github.com/Painkiller675/url_shortener_6750/internal/middleware/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	// init config
	config.SetConfig()

	// init logger
	initLogger()

	// init router
	r := chi.NewRouter()

	// init storage if any

	// set logger for chi router
	r.Use(logger.LogMW)
	r.Use(gzipMW.GzMW)

	// routing
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.CreateShortURLHandler)
		r.Get("/{id}", handlers.GetLongURLHandler)
		r.Post("/api/shorten", handlers.CreateShortURLJSONHandler)
	})
	//start server
	logger.Log.Info("Running server", zap.String("address", config.StartOptions.HTTPServer.Address))
	if err := http.ListenAndServe(config.StartOptions.HTTPServer.Address, r); err != nil {
		panic(err)
	}

}

func initLogger() {
	if err := logger.Initialize(config.StartOptions.LogLvl); err != nil {
		panic(err) // TODO How to handle it??
	}
}
