package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/middleware/logger"
	"github.com/Painkiller675/url_shortener_6750/internal/repository"
	"github.com/Painkiller675/url_shortener_6750/internal/service"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func CreateShortURLJSONHandler(res http.ResponseWriter, req *http.Request) {

	//err := json.NewDecoder(resp.Body).Decode(&jsonData)
	//fmt.Println("body = ", body)
	//check content-type
	if ok := strings.Contains(req.Header.Get("Content-Type"), "application/json"); !ok {
		logger.Log.Info("[INFO]", zap.String("body", "no content type"), zap.String("method", req.Method), zap.String("url", req.URL.Path))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var jsStruct repository.JSONStructSh
	var orStruct repository.JSONStructOr
	var buf bytes.Buffer
	// feed data from the body into the buffer
	if _, err := buf.ReadFrom(req.Body); err != nil {
		logger.Log.Info("[ERROR]", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// deserialize JSON into JSStruct
	if err := json.Unmarshal(buf.Bytes(), &orStruct); err != nil {
		logger.Log.Info("[ERROR]", zap.Error(err))
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("orStruct.OrURL = ", orStruct.OrURL)
	// calculate the alias
	randAl := service.GetRandString(8)
	// write into safeStorage to allow getting the data
	repository.WriteURL(randAl, orStruct.OrURL)
	// base URL
	baseURL := config.StartOptions.BaseURL
	shURL, err := url.JoinPath(baseURL, randAl)

	if err != nil {
		logger.Log.Info("[ERROR]", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// add short URL to the auxiliary struct
	jsStruct.ShURL = shURL
	// marshal data for response
	marData, err := json.Marshal(jsStruct)
	if err != nil {
		logger.Log.Info("[ERROR]", zap.Error(err))
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	// headers molding ..
	res.Header().Set("Content-Type", "application"+
		"/json")
	res.Header().Set("Content-Length", strconv.Itoa(len(marData)))
	res.WriteHeader(http.StatusCreated) // 201
	// response body molding
	_, err = res.Write(marData)
	if err != nil {
		logger.Log.Info("[ERROR]", zap.Error(err))
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

}
