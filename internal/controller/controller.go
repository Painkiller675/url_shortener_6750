package controller

import (
	"bytes"
	"encoding/json"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/repository"
	"github.com/Painkiller675/url_shortener_6750/internal/service"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Controller struct {
	logger  *zap.Logger
	storage *repository.Storage
}

func New(logger *zap.Logger, storage *repository.Storage) *Controller {
	return &Controller{logger: logger, storage: storage}
}

func (c *Controller) CreateShortURLHandler(res http.ResponseWriter, req *http.Request) {

	body, err := io.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// write an alias
	randAl := service.GetRandString(8)
	c.storage.SafeStorage.StoreAlURL(randAl, string(body))
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
		c.logger.Info("Failed to write response", zap.Error(err))
		return
	}
}

func (c *Controller) GetLongURLHandler(res http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id") // the cap
	// response molding ...
	orURL, err := c.storage.SafeStorage.GetOrURL(id)
	if err != nil { // TODO: mb I should use status 500 here?
		c.logger.Info("Failed to get orURL", zap.String("id", id), zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Location", orURL)
	res.WriteHeader(http.StatusTemporaryRedirect) // 307
}

func (c *Controller) CreateShortURLJSONHandler(res http.ResponseWriter, req *http.Request) {

	//check content-type
	if ok := strings.Contains(req.Header.Get("Content-Type"), "application/json"); !ok {
		c.logger.Info("[INFO]", zap.String("body", "no content type"), zap.String("method", req.Method), zap.String("url", req.URL.Path))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var jsStruct repository.JSONStructSh
	var orStruct repository.JSONStructOr
	var buf bytes.Buffer
	// feed data from the body into the buffer
	if _, err := buf.ReadFrom(req.Body); err != nil {
		c.logger.Info("[ERROR]", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// deserialize JSON into JSStruct
	if err := json.Unmarshal(buf.Bytes(), &orStruct); err != nil {
		c.logger.Info("[ERROR]", zap.Error(err))
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	// calculate the alias
	randAl := service.GetRandString(8)
	// write into safeStorage to allow getting the data
	c.storage.SafeStorage.StoreAlURL(randAl, orStruct.OrURL)
	// base URL operations
	baseURL := config.StartOptions.BaseURL
	shURL, err := url.JoinPath(baseURL, randAl)

	if err != nil {
		c.logger.Info("[ERROR]", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// add short URL to the auxiliary struct
	jsStruct.ShURL = shURL
	// marshal data for response
	marData, err := json.Marshal(jsStruct)
	if err != nil {
		c.logger.Info("[ERROR]", zap.Error(err))
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	// headers molding
	res.Header().Set("Content-Type", "application"+
		"/json")
	res.WriteHeader(http.StatusCreated) // 201
	// response body molding
	_, err = res.Write(marData)
	if err != nil {
		c.logger.Info("[ERROR]", zap.Error(err))
		return
	}

}
