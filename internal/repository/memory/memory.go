// Memory package is used to implement a memory database logic in the app.
package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/Painkiller675/url_shortener_6750/internal/lib/merrors"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
)

// Storage - the instance of memory database type.
type Storage struct {
	storage *[]storageWithUserID
	logger  *zap.Logger
}

// NewStorage is a constructor of a memory database.
func NewStorage(logger *zap.Logger) *Storage {
	logger.Info("MEMORY storage is available")
	return &Storage{logger: logger, storage: &[]storageWithUserID{}}
	//return &Storage{storage: &[]storageWithUserID{}}
}

type storageWithUserID struct {
	UserID   string            `json:"-"` // TODO: use lower case
	alURLMap map[string]string `json:"-"`
	mx       *sync.RWMutex     `json:"-"`
}

// newStorageWithUserID - constructor (embedding) to implement the user-accessory and delete flags.
func newStorageWithUserID(userID string, alias string, url string) *storageWithUserID {
	// 1st init
	var alurlmap = make(map[string]string)
	alurlmap[alias] = url
	return &storageWithUserID{
		UserID:   userID,
		alURLMap: alurlmap, // TODO [MENTOR] Do I :assign the copy here? Is it OK? I get it by still ..
		mx:       &sync.RWMutex{}}

}

// StoreAlURL is used for storing alias, url and userID in a memory database.
func (s *Storage) StoreAlURL(_ context.Context, alias string, url string, userID string) (int64, error) {
	// find needed storage for a specific userID
	for _, userIDStorage := range *s.storage {
		if userIDStorage.UserID == userID {
			userIDStorage.mx.Lock() // lock data for the map for a specific userID
			defer userIDStorage.mx.Unlock()
			userIDStorage.alURLMap[alias] = addExistMarker(url)
			// we have handled it for userID needed => break
			return 1, nil
		}
	}
	// new user init
	// We haven't found the user with userID in the storage => add new userID
	*(s.storage) = append(*(s.storage), *newStorageWithUserID(userID, alias, addExistMarker(url)))
	return 1, nil

}

// GetOrURLByAl returns #1st found# original url notwithstanding what user created it
func (s *Storage) GetOrURLByAl(_ context.Context, alias string) (string, error) {
	const op = "memory.GetOrURLByAl"
	// Если горутина собирается читать данные, то она вызывает метод RLock(). Метод RLock() не
	//даёт начать запись пока не будут завершены все операции чтения.
	if s.storage == nil { // TODO [MENTOR] IS IT NEEDED???!
		er := fmt.Errorf("original URL for %v doesn't exist in the memory-DB", alias)
		return "", er
	}
	for _, userIDStorage := range *s.storage {
		userIDStorage.mx.RLock()
		defer userIDStorage.mx.RUnlock()
		if orURL, ok := userIDStorage.alURLMap[alias]; ok {
			if isExist(orURL) { // CHECK the 1st founded URL !
				return delMarker(orURL), nil
			}
			// URL doesn't exist!
			return "", fmt.Errorf("%s: %w", op, merrors.ErrURLNotFound)
		}
	}
	er := fmt.Errorf("original URL for %v doesn't exist in the DB", alias)
	return "", er

}

// GetDataByUserID gets short URLs and original URLs of a particular user.
func (s *Storage) GetDataByUserID(ctx context.Context, userID string) (*[]models.UserURLS, error) {
	if s.storage == nil {
		er := fmt.Errorf("no data for %v", userID)
		return nil, er
	}
	for _, userIDStorage := range *s.storage {
		if userIDStorage.UserID == userID {
			var dataOfuserID = make([]models.UserURLS, 0, 20)
			userIDStorage.mx.RLock()
			defer userIDStorage.mx.RUnlock()
			for al, url := range userIDStorage.alURLMap {
				dataOfuserID = append(dataOfuserID, models.UserURLS{ShortURL: al, OriginalURL: delMarker(url)})
			}
			return &dataOfuserID, nil

		}
	}
	// such a userID doesn't have any record
	er := fmt.Errorf("no data for %v", userID)
	return nil, er

}

// SaveBatchURL saves the batch of URLs in a memory database.
func (s *Storage) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	const op = "memory.SaveBatchURL"
	// create the arrays of structs for response
	toResp := make([]models.JSONBatStructToSerResp, 0) // TODO [MENTOR]: is it ok allocation? why len(*corURLSh) is false? (instead of 0)
	// saving ..
	for _, idURLSh := range *corURLSh {
		_, err := s.StoreAlURL(ctx, idURLSh.ShortURL, idURLSh.OriginalURL, "") // TODO: how to use _ here?
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

// Ping checks if postgreSQL database is available HERE it's a blind plug.
func (s *Storage) Ping(ctx context.Context) error {
	return errors.New("DB isn't available")
}

// GetAlByURL is a blind plug here.
func (s *Storage) GetAlByURL(ctx context.Context, url string) (string, error) { return "", nil }

// DeleteURLsByUserID deletes some records from the database by UserID (set flag is_deleted in true state)
// the func doesn't use transaction just to implement  a multi stream concept.
func (s *Storage) DeleteURLsByUserID(_ context.Context, userID string, aliasToDel []string) error {
	const op = "memory.DeleteURLsByUserID"
	if s.storage == nil {
		return fmt.Errorf("[INFO] memory DB is empty")
	}
	for _, userIDStorage := range *s.storage {
		if userIDStorage.UserID == userID {
			userIDStorage.mx.Lock()
			defer userIDStorage.mx.Unlock()
			for _, alToDel := range aliasToDel {
				// if we have such alias in the memory => del this
				if _, ok := userIDStorage.alURLMap[alToDel]; ok {
					userIDStorage.alURLMap[alToDel] = changeExistToDelMarker(userIDStorage.alURLMap[alToDel])
				}
			}
			return nil
		}

	}
	// if we don't have such a user in the DB
	return fmt.Errorf("[%v] user doesn't exist", op)
}

// CheckIfUserExists checks the existence of a user.
func (s *Storage) CheckIfUserExists(_ context.Context, userID string) error {
	const op = "memory.CheckIfUserExists"
	for _, userIDStorage := range *s.storage {
		if userIDStorage.UserID == userID {
			return nil
		}
	}
	return fmt.Errorf("[%s]: %w", op, merrors.ErrUserNotFound)
}

// functions to implement deleting
// delMarker deletes the last symbol
func delMarker(s string) string {
	s = s[:len(s)-1]
	return s
}
func addExistMarker(s string) string {
	s = s + "@"
	return s
}

func changeExistToDelMarker(s string) string {
	s = s[:len(s)-1]
	s = s + "-"
	return s
}

func isExist(s string) bool {
	return s[len(s)-1:] == "@"
}
