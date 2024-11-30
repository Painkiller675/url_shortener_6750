package handlers

import (
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/repository"
	"github.com/Painkiller675/url_shortener_6750/internal/service"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func CreateShortURLHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// write an alias
	randAl := service.GetRandString(8)
	repository.WriteURL(randAl, string(body))
	// response molding
	baseURL := config.StartOptions.BaseURL
	resultURL, err := url.JoinPath(baseURL, randAl)
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Content-Length", strconv.Itoa(len([]byte(resultURL))))
	res.WriteHeader(http.StatusCreated) // 201
	_, err = res.Write([]byte(resultURL))
	if err != nil {
		log.Printf("Error writing to response: %v", err)
		return
	}
}

func GetLongURLHandler(res http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id") // the cap
	// response molding ...
	orURL, err := repository.GetShortURL(id)
	if err != nil { // TODO: mb I should use status 500 here?
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", orURL)
	res.WriteHeader(http.StatusTemporaryRedirect) // 307
}
