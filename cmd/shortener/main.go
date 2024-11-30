package main

import (
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	// init config
	config.SetConfig()

	// init router
	r := chi.NewRouter()

	// routing
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.CreateShortURLHandler)
		r.Get("/{id}", handlers.GetLongURLHandler)
	})

	//start server
	err := http.ListenAndServe(config.StartOptions.HTTPServer.Address, r)
	if err != nil {
		panic(err) // or log.Fatal()???

	} // TODO: How should I handle the error over here?
}
