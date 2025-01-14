package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// JSONStruct is used to unmarshal js request nd send js response in CreateShortURLJSONHandler
type JSONStructSh struct {
	ShURL string `json:"result"`
}
type JSONStructOr struct {
	OrURL string `json:"url"`
}

// structs for the batch

type JSONBatStructDes struct {
	CorrelationID int64  `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type JSONBatStructSer struct {
	CorrelationID int64  `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Controller struct {
	logger  *zap.Logger
	storage repository.URLStorage
}

func New(logger *zap.Logger, storage repository.URLStorage) *Controller {
	return &Controller{logger: logger, storage: storage}
}

func (c *Controller) CreateShortURLHandler(ctx context.Context) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		const op = "controller.CreateSHortURLHandler"
		body, err := io.ReadAll(req.Body)
		//check the body
		if err != nil || len(body) == 0 {
			c.logger.Info("Body is empty!", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// write an alias
		randAl := service.GetRandString(8)
		_, err = c.storage.StoreAlURL(ctx, randAl, string(body)) // TODO [MENTOR]: mb del _ or change driver to support id?
		if err != nil {
			c.logger.Info("Failed to store URL", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// response molding
		baseURL := config.StartOptions.BaseURL
		fmt.Println("BASE URL in CreateNoJS = ", baseURL)
		resultURL, err := url.JoinPath(baseURL, randAl)
		fmt.Println("resultURL in CreateNOJS = ", resultURL)
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
}

func (c *Controller) GetLongURLHandler(ctx context.Context) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		idAl := req.PathValue("id") // the cap
		// response molding ...
		orURL, err := c.storage.GetOrURLByAl(ctx, idAl)
		if err != nil { // TODO: mb I should use status 500 here?
			c.logger.Info("Failed to get orURL", zap.String("id", idAl), zap.Error(err))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", orURL)
		res.WriteHeader(http.StatusTemporaryRedirect) // 307
	}
}

func (c *Controller) CreateShortURLJSONHandler(ctx context.Context) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		const op = "controller.CreateShortURLJSONHandler"
		//check content-type
		if ok := strings.Contains(req.Header.Get("Content-Type"), "application/json"); !ok {
			c.logger.Info("[INFO]", zap.String("body", "no content type"), zap.String("method", req.Method), zap.String("url", req.URL.Path))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		var jsStruct JSONStructSh
		var orStruct JSONStructOr
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
		// write into a storage to allow getting the data
		_, err := c.storage.StoreAlURL(ctx, randAl, orStruct.OrURL)
		if err != nil {
			c.logger.Info("Failed to store URL", zap.String("place:", op), zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// base URL
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
		//res.Header().Set("Content-Length", strconv.Itoa(len(marData)))
		res.WriteHeader(http.StatusCreated) // 201
		// response body molding
		_, err = res.Write(marData)
		if err != nil {
			c.logger.Info("[ERROR]", zap.Error(err))
			return
		}
	}
}

func (c *Controller) PingDB(ctx context.Context) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//ctx := context.Background()
		err := c.storage.Ping(ctx)
		// if no connection
		if err != nil {
			c.logger.Info("[ERROR]", zap.String("PingDB", "Can't ping pg database!"), zap.Error(err))
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		// if connected
		res.WriteHeader(http.StatusOK)

	}
}

/*
func (c *Controller) CreateShortURLJSONBatchHandler (res http.ResponseWriter, req *http.Request) {
		//check content-type
		if ok := strings.Contains(req.Header.Get("Content-Type"), "application/json"); !ok {
			c.logger.Info("[INFO]", zap.String("body", "no content type"), zap.String("method", req.Method), zap.String("url", req.URL.Path))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		var jsStructsDes []JSONBatStructDes
		var jsStructsSer []JSONBatStructSer
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
		fmt.Println("orStruct.OrURL = ", orStruct.OrURL)
		// calculate the alias
		randAl := service.GetRandString(8)
		// write into safeStorage to allow getting the data
		c.storage.SafeStorage.StoreAlURL(randAl, orStruct.OrURL)
		// base URL
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
		//res.Header().Set("Content-Length", strconv.Itoa(len(marData)))
		res.WriteHeader(http.StatusCreated) // 201
		// response body molding
		_, err = res.Write(marData)
		if err != nil {
			c.logger.Info("[ERROR]", zap.Error(err))
			return
		}
}
*/
