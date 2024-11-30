package main

import (
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/handlers"
	"net/http"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handlers.CreateShortURLHandler)
	mux.HandleFunc("GET /{id}", handlers.GetLongURLHandler)

	//start server
	err := http.ListenAndServe(cfg.Address, mux)
	if err != nil {
		panic(err) // or log.Fatal()???

	}
}
