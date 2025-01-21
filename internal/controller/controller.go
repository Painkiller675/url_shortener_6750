package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/lib/merrors"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
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
		randAl := service.GetRandString(string(body))
		_, err = c.storage.StoreAlURL(ctx, randAl, string(body)) // TODO [MENTOR]: mb del _ or change driver to support id?
		httpStatus := http.StatusCreated
		if err != nil {
			if errors.Is(err, merrors.ErrURLOrAliasExists) { // the try to short already existed url pg database
				c.logger.Info("URL already exists!", zap.Error(err))
				httpStatus = http.StatusConflict
			} else {
				c.logger.Info("Failed to store URL", zap.Error(err))
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

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
		res.WriteHeader(httpStatus) // 201 or 409
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
		fmt.Println("orURL in GetLongURLHandler = ", orURL)
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
		randAl := service.GetRandString(orStruct.OrURL)
		// write into a storage to allow getting the data
		_, err := c.storage.StoreAlURL(ctx, randAl, orStruct.OrURL)
		httpStatus := http.StatusCreated
		if err != nil {
			if errors.Is(err, merrors.ErrURLOrAliasExists) { // if alias for url already exists in the pg database
				c.logger.Info("URL already exists!", zap.Error(err))
				httpStatus = http.StatusConflict
			} else {
				c.logger.Info("Failed to store URL", zap.String("place:", op), zap.Error(err))
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
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
		res.Header().Set("Content-Type", "application/json")
		//res.Header().Set("Content-Length", strconv.Itoa(len(marData)))
		res.WriteHeader(httpStatus) // 201 or 409
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
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// if connected
		res.WriteHeader(http.StatusOK)

	}
}

func (c *Controller) CreateShortURLJSONBatchHandler(ctx context.Context) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//check content-type (application/json)
		if ok := strings.Contains(req.Header.Get("Content-Type"), "application/json"); !ok {
			c.logger.Info("[INFO]", zap.String("body", "no content type application/json"), zap.String("method", req.Method), zap.String("url", req.URL.Path))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// check the body: TODO: del reuse bbody

		//check the body
		body, err := io.ReadAll(req.Body)
		if err != nil || len(body) == 0 {
			c.logger.Info("Body is empty!", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Replace the body with a new reader after reading from the original
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		// create the array of structures to deserialize data
		var desBatchStruct []models.JSONBatStructToDesReq

		//TODO: [MENTOR] Should I implement such a check?
		/*
			dec := json.NewDecoder(req.Body)
				if err := dec.Decode(&request); err != nil {
					httpError.RespondWithError(res, http.StatusInternalServerError, "Invalid JSON body")
					return
				}
		*/

		var buf bytes.Buffer
		// feed data from the body into the buffer
		if _, err := buf.ReadFrom(req.Body); err != nil {
			c.logger.Info("[ERROR]", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest) // TODO [MENTOR]: BadRequest or InternalServerError?
			return
		}
		defer req.Body.Close() // TODO [MENTOR] I didn't assign it should I close it??
		// deserialize JSON batch into desBatchStruct
		if err := json.Unmarshal(buf.Bytes(), &desBatchStruct); err != nil {
			c.logger.Info("[ERROR]", zap.Error(err))
			http.Error(res, err.Error(), http.StatusBadRequest) // TODO [MENTOR]: BadRequest or InternalServerError?
			return
		}
		fmt.Println("desBatchStruct = ", desBatchStruct)

		// create an auxiliary array of structures
		idURLAl, err := service.CreateBatchIDOrSh(&desBatchStruct)
		fmt.Println("idURLAl = ", idURLAl)
		if err != nil {
			c.logger.Info("[INFO]", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // TODO [MENTOR]: BadRequest or InternalServerError?
			return
		}

		// save data into the database and create respBatch for response
		respBatch, err := c.storage.SaveBatchURL(ctx, idURLAl)
		fmt.Println("respBatch = ", *respBatch)
		if err != nil {
			c.logger.Info("[ERROR]", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// create the array of structs to add base url to response
		response := make([]models.JSONBatStructToSerResp, 0) //TODO [MENTOR] is it a good allocation or I should use len?
		// molding the response

		for _, idSh := range *respBatch {
			fullShortURL, err := url.JoinPath(config.StartOptions.BaseURL, idSh.ShortURL)
			if err != nil {
				c.logger.Info("[ERROR]", zap.Error(err))
			}
			response = append(response, models.JSONBatStructToSerResp{
				CorrelationID: idSh.CorrelationID,
				ShortURL:      fullShortURL,
			})
		}

		// marshal data for response
		marData, err := json.Marshal(response)
		if err != nil {
			c.logger.Info("[ERROR]", zap.Error(err))
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		// write headers
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated) // 201

		// response body molding
		_, err = res.Write(marData)
		if err != nil {
			c.logger.Info("[ERROR]", zap.Error(err))
			return
		}
	}
}
