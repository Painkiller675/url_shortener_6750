// the package is for getting the access to the business logic via Business interface
package business

import (
	"context"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"github.com/Painkiller675/url_shortener_6750/internal/service"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"

	"github.com/Painkiller675/url_shortener_6750/internal/repository"
)

// Business is used to implement all the methods with the business logic
type Business struct {
	Storage repository.URLStorage
	logger  *zap.Logger
}

func NewBusiness(storage repository.URLStorage, logger *zap.Logger) *Business {
	return &Business{Storage: storage, logger: logger}
}

// CheckIfUserExists checks if a user exists in the database. It's used in DeleteURLSHandler handler (GET /api/user/urls)
func (b *Business) CheckIfUserExists(ctx context.Context, userID string) error {
	return b.Storage.CheckIfUserExists(ctx, userID)
}

// GetStats is used to get the statistics namely the number of urls and users in the database - /api/internal/stats
func (b *Business) GetStats(ctx context.Context) (urls int, users int, err error) {
	return b.Storage.GetStats(ctx)
}

// PingDB is used to figure out if the database is available
func (b *Business) PingDB(ctx context.Context) error {
	return b.Storage.Ping(ctx)
}

// StoreAlURL is used for / handler it returns molded URL and an error if any
func (b *Business) StoreAlURL(ctx context.Context, urlIn *url.URL, userID string) (string, error) {
	//errs := make([]error, 0)
	urlInStr := urlIn.String()                // TODO: привести к стрингу тут или слать приведение везде?
	randAl := service.GetRandString(urlInStr) // TODO: mb replace it to business
	_, err := b.Storage.StoreAlURL(ctx, randAl, urlInStr, userID)
	//errs = append(errs, err)
	baseURL := config.StartOptions.BaseURL
	baseURL = baseURL.JoinPath(randAl) // resultURL // TODO: reassign??
	//resultURL, err := url.JoinPath(baseURL, randAl) // err == nil anyway because we have already parsed the url in the  handler
	//errs = append(errs, err)
	return baseURL.String(), err // I return resultURL anyway
}

// GetOrURLByAl is used for /{id}  (in REST controller - GetLongURLHandler handler)
func (b *Business) GetOrURLByAl(ctx context.Context, alias string) (string, error) {
	return b.Storage.GetOrURLByAl(ctx, alias)
}

// SaveBatchURL is used in /api/shorten/batch (REST CreateShortURLJSONBatchHandler handler)
func (b *Business) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	return b.Storage.SaveBatchURL(ctx, corURLSh)
}

// GetDataByUserID  /api/user/urls returns aliases of a particular user (GetUserURLSHandler handler)
func (b *Business) GetDataByUserID(ctx context.Context, userID string) (*[]models.UserURLS, error) {
	return b.Storage.GetDataByUserID(ctx, userID)
}

// TODO: small case ??
// retrieveUserIDFromTokenString retrieves userID from a token string
func (b *Business) RetrieveUserIDFromTokenString(tokenStr string) (string, error) {

	// создаём экземпляр структуры с утверждениями
	claims := &models.Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		} // anti-hacker check
		return []byte(config.SecretKey), nil
	})
	if err != nil {
		b.logger.Info("Can't parse token!", zap.Error(err))
		return "", errors.New("can't parse token")
	}

	if !token.Valid {
		b.logger.Info("Invalid token!", zap.Error(err))
		return "", errors.New("invalid token")
	}

	b.logger.Info("Successfully retrieved token!", zap.String("token", tokenStr))
	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, nil

}

// GetTokenStrVal returns tokenStr.Value from the Cookies
func (b *Business) GetTokenStrVal(req *http.Request) (string, error) {
	// get token string from the cookies
	tokenString, err := req.Cookie("token")
	if err != nil {
		b.logger.Info("no token", zap.Error(err))
		return "", errors.New("no token")
	}
	// Check token value and send it for retrieving userID
	if tokenString.Value == "" {
		b.logger.Info("empty token", zap.Error(err))
		return "", errors.New("empty token")
	}
	return tokenString.Value, nil
}

func (b *Business) SetAuthTokenInCookies(w http.ResponseWriter, tokenStr string) {

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenStr,
		Secure:   false,
		HttpOnly: true,
		Expires:  time.Now().Add(config.TokenExp),
	})

}

// genJWTTokenString create JWT token and return it in string type.
func (b *Business) GenJWTTokenString() (string, string, error) { // TODO [MENTOR]: mb I should replace this func ???
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
