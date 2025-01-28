package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/lib/merrors"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"github.com/Painkiller675/url_shortener_6750/internal/repository"
	"github.com/Painkiller675/url_shortener_6750/internal/service"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

// genJWTTokenString create JWT token and return it in string type
func (c *Controller) genJWTTokenString() (string, string, error) { // TODO [MENTOR]: mb I should replace this func ???
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	//usId := string(time.Now().Unix())
	usID := service.GetRandString(time.Now().UTC().String())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// set expiration time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenExp)), //TODO [MENTOR] is it a good way to store it?
		},
		// set my own statement
		UserID: usID, // TODO [MENTOR]: how should I implement it better??
		// int(b[0] + b[1])
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(config.SecretKey)) // TODO [MENTOR]: how to store it better? how people store it in real projects? In env?
	// TODO: ok if env .. I set the env value secretKey on my PC e.g. and then start the app?
	if err != nil {
		return "", "", err
	}

	// возвращаем строку токена
	return tokenString, usID, nil

}

func (c *Controller) retrieveUserIDFromTokenString(r *http.Request) string { // TODO [MENTOR]: mb I should replace this func ???
	// get token string from the cookies
	tokenString, err := r.Cookie("token")

	if err != nil {
		c.logger.Info("No token!", zap.Error(err))
		return "-1"
	}
	// TODO: [MENTOR] SHOULD I CHECK
	if tokenString.Value == "" {
		c.logger.Info("Empty token!", zap.Error(err))
		return "-1"
	}
	// создаём экземпляр структуры с утверждениями
	claims := &models.Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString.Value, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		} // anti-hacker check
		return []byte(config.SecretKey), nil
	})
	if err != nil {
		c.logger.Info("Can't parse token!", zap.Error(err))
		return "-1"
	}

	if !token.Valid {
		c.logger.Info("Invalid token!", zap.Error(err))
		return "-1"
	}

	c.logger.Info("Successfully retrieved token!", zap.String("token", tokenString.Value))
	// возвращаем ID пользователя в читаемом виде
	return claims.UserID

}

func (c *Controller) setAuthToken(w http.ResponseWriter, tokenStr string) {

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenStr,
		Expires: time.Now().Add(config.TokenExp),
	})

}

func (c *Controller) CreateShortURLHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		const op = "controller.CreateSHortURLHandler"

		body, err := io.ReadAll(req.Body)
		//check the body
		if err != nil || len(body) == 0 {
			c.logger.Info("Body is empty!", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var tokenStr, userID string

		// retrieve token if any
		userID = c.retrieveUserIDFromTokenString(req)
		if userID == "-1" { // can't retrieve => register a new user a
			tokenStr, userID, err = c.genJWTTokenString()
			if err != nil {
				c.logger.Info("Can't generate token!", zap.Error(err))
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		// add token string to the Cookies
		c.setAuthToken(res, tokenStr)
		// save the data
		randAl := service.GetRandString(string(body))
		_, err = c.storage.StoreAlURL(req.Context(), randAl, string(body), userID) // TODO [MENTOR]: mb del _ or change driver to support id?
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
		resultURL, err := url.JoinPath(baseURL, randAl)
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

func (c *Controller) GetLongURLHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		idAl := req.PathValue("id") // the cap
		// response molding ...
		orURL, err := c.storage.GetOrURLByAl(req.Context(), idAl)
		if err != nil { // TODO: mb I should use status 500 here?
			c.logger.Info("Failed to get orURL", zap.String("id", idAl), zap.Error(err))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", orURL)
		res.WriteHeader(http.StatusTemporaryRedirect) // 307
	}
}

func (c *Controller) CreateShortURLJSONHandler() http.HandlerFunc {
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

		var tokenStr, userID string
		var err error
		// retrieve token if any
		userID = c.retrieveUserIDFromTokenString(req)
		if userID == "-1" { // can't retrieve => register a new user a
			tokenStr, userID, err = c.genJWTTokenString()
			if err != nil {
				c.logger.Info("Can't generate token!", zap.Error(err))
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		// add token string to the Cookies
		c.setAuthToken(res, tokenStr)

		// calculate the alias
		randAl := service.GetRandString(orStruct.OrURL)
		// save the data
		_, err = c.storage.StoreAlURL(req.Context(), randAl, orStruct.OrURL, userID)
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
			c.logger.Error("[ERROR]", zap.Error(err))
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
			c.logger.Error("[ERROR]", zap.Error(err))
			return
		}
	}
}

func (c *Controller) PingDB() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := c.storage.Ping(req.Context())
		// if no connection
		if err != nil {
			c.logger.Warn("[WARNING]", zap.String("PingDB", "Can't ping pg database!"), zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// if connected
		res.WriteHeader(http.StatusOK)

	}
}

func (c *Controller) CreateShortURLJSONBatchHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//check content-type (application/json)
		if ok := strings.Contains(req.Header.Get("Content-Type"), "application/json"); !ok {
			c.logger.Warn("[WARNING]", zap.String("body", "no content type application/json"), zap.String("method", req.Method), zap.String("url", req.URL.Path))
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
			c.logger.Error("[ERROR]", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest) // TODO [MENTOR]: BadRequest or InternalServerError?
			return
		}
		defer req.Body.Close() // TODO [MENTOR] I didn't assign it should I close it??
		// deserialize JSON batch into desBatchStruct
		if err := json.Unmarshal(buf.Bytes(), &desBatchStruct); err != nil {
			c.logger.Error("[ERROR]", zap.Error(err))
			http.Error(res, err.Error(), http.StatusBadRequest) // TODO [MENTOR]: BadRequest or InternalServerError?
			return
		}

		// create an auxiliary array of structures
		idURLAl, err := service.CreateBatchIDOrSh(&desBatchStruct)
		if err != nil {
			c.logger.Error("[ERROR]", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // TODO [MENTOR]: BadRequest or InternalServerError?
			return
		}

		// save data into the database and create respBatch for response
		respBatch, err := c.storage.SaveBatchURL(req.Context(), idURLAl)
		if err != nil {
			c.logger.Error("[ERROR]", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// create the array of structs to add base url to response
		response := make([]models.JSONBatStructToSerResp, 0, len(*respBatch)) //
		// molding the response

		for _, idSh := range *respBatch {
			fullShortURL, err := url.JoinPath(config.StartOptions.BaseURL, idSh.ShortURL)
			if err != nil {
				c.logger.Error("[ERROR]", zap.Error(err))
			}
			response = append(response, models.JSONBatStructToSerResp{
				CorrelationID: idSh.CorrelationID,
				ShortURL:      fullShortURL,
			})
		}

		// marshal data for response
		marData, err := json.Marshal(response)
		if err != nil {
			c.logger.Error("[ERROR]", zap.Error(err))
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		// write headers
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated) // 201

		// response body molding
		_, err = res.Write(marData)
		if err != nil {
			c.logger.Error("[ERROR]", zap.Error(err))
			return
		}
	}
}

func (c *Controller) GetUserURLSHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		const op = "controller.GetUserURLSHandler"

		// retrieve token if any
		userID := c.retrieveUserIDFromTokenString(req)
		if userID == "-1" { // can't retrieve => return 401 Unauthorized
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		//var alURLStruct = models.UserURLS{}
		respAlURLStruct, err := c.storage.GetDataByUserID(req.Context(), userID)
		if err != nil {
			if errors.Is(err, merrors.ErrURLNotFound) { // no data for the user!
				c.logger.Info("[INFO]", zap.String("place:", op), zap.Error(err))
				res.WriteHeader(http.StatusNoContent)
				return
			}
			// handle other possible errors
			c.logger.Error("[ERROR]", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// replace alias with short url (add base url)
		for n, alURL := range *respAlURLStruct {
			fullShortURL, err := url.JoinPath(config.StartOptions.BaseURL, alURL.ShortURL)
			if err != nil {
				c.logger.Error("[ERROR]", zap.Error(err))
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			(*(respAlURLStruct))[n].ShortURL = fullShortURL

		}

		// marshal data for response
		marData, err := json.Marshal(*respAlURLStruct)
		if err != nil {
			c.logger.Error("[ERROR]", zap.Error(err))
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		// write headers
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK) // 200

		// response body molding
		_, err = res.Write(marData)
		if err != nil {
			c.logger.Error("[ERROR]", zap.Error(err))
			return

		}
	}
}
