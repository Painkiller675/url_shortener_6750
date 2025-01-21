package memory

import (
	"context"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/models"

	//"github.com/Painkiller675/url_shortener_6750/internal/config"
	"go.uber.org/zap"
	"sync"
)

/*type Storage struct {
	SafeStorage *safeStorage
	logger      *zap.Logger
	//filename string
}

func NewStorage(log *zap.Logger) *Storage {
	return &Storage{
		logger:      log,
		SafeStorage: newSafeStorage(log),
		//filename: config.StartOptions.Filename,
	}
}
*/
// TODO constructor safeStorage
type Storage struct {
	AlURLStorage map[string]string `json:"al_url_storage"`
	mx           *sync.RWMutex     `json:"-"` // TODO pointer or not??
	logger       *zap.Logger       `json:"-"`
}

func NewStorage(logger *zap.Logger) *Storage {
	return &Storage{
		AlURLStorage: make(map[string]string), // mb save all the struct but wht about logger etc?
		mx:           &sync.RWMutex{},
		logger:       logger,
	}
}

func (s *Storage) StoreAlURL(_ context.Context, alias string, url string) (int64, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.AlURLStorage[alias] = url //TODO: mb I should somehow handle that?
	return 1, nil               // blind plug
}

func (s *Storage) GetOrURLByAl(_ context.Context, alias string) (string, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if orURL, ok := s.AlURLStorage[alias]; ok {
		return orURL, nil
	}
	er := fmt.Errorf("original URL for %v doesn't exist in the DB", alias)
	s.logger.Info("[INFO]", zap.Error(er))
	return "", er //TODO: handle that more properly
}

func (s *Storage) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	const op = "memory.SaveBatchURL"
	// create the arrays of structs for response
	toResp := make([]models.JSONBatStructToSerResp, 0) // TODO [MENTOR]: is it ok allocation? why len(*corURLSh) is false? (instead of 0)
	// saving ..
	for _, idURLSh := range *corURLSh {
		_, err := s.StoreAlURL(ctx, idURLSh.ShortURL, idURLSh.OriginalURL) // TODO: how to use _ here?
		if err != nil {
			s.logger.Info(op, zap.Error(err))
			return nil, err
		}
		// molding object for response
		toResp = append(toResp, models.JSONBatStructToSerResp{
			CorrelationID: idURLSh.CorrelationID,
			ShortURL:      idURLSh.ShortURL,
		})
	}
	return &toResp, nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return errors.New("DB isn't available")
}

func (s *Storage) GetAlByURL(ctx context.Context, url string) (string, error) { return "", nil }
