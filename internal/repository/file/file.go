// File package is used to implement a file database logic in the app.
package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/lib/merrors"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
)

// Storage is a basic struct of a file storage
type Storage struct {
	//AlURLStorage map[string]string `json:"url_storage"`
	AlURLStorage []*storageWithUserID `json:"url_storage"`
	Filename     string               `json:"-"`
	logger       *zap.Logger          `json:"-"`
	//mx           *sync.RWMutex     `json:"-"` // TODO pointer or not??
}

// NewStorage is a constructor of a memory database.
func NewStorage(filename string, logger *zap.Logger) *Storage {
	logger.Info("FILE storage is available")
	//feed data from the file into the memory
	stor, err := getStorage(filename)
	if err != nil {
		logger.Fatal("[FATAL] file storage is not available", zap.Error(err))
	}
	return &Storage{
		AlURLStorage: stor.AlURLStorage, // mb save all the struct but wht about logger etc?
		//mx:           &sync.RWMutex{},
		Filename: filename,
		logger:   logger,
	}
}

type storageWithUserID struct {
	UserID       string            `json:"userID"`
	AlURLStorage map[string]string `json:"url_storage"`
	mx           *sync.RWMutex     `json:"-"`
}

// newStorageWithUserID - constructor (embedding) to implement the user-accessory and delete flags.
func newStorageWithUserID(userID string, alias string, url string) *storageWithUserID {
	// 1st init
	var alurlmap = make(map[string]string)
	alurlmap[alias] = url
	return &storageWithUserID{
		UserID:       userID,
		AlURLStorage: alurlmap, // TODO [MENTOR] Do I :assign the copy here? Is it OK? I get it by still ..
		mx:           &sync.RWMutex{}}

}

// Close is a blind plug here cause we close file in save & get functions
func (s *Storage) Close() error {
	return nil
}

// StoreAlURL is used for storing alias, url and userID in a file database.
func (s *Storage) StoreAlURL(_ context.Context, alias string, orURL string, userID string) (int64, error) {
	// find needed storage for a specific userID
	for _, userIDStor := range s.AlURLStorage {
		if userIDStor.UserID == userID {
			userIDStor.mx.Lock()
			defer userIDStor.mx.Unlock()
			userIDStor.AlURLStorage[alias] = addExistMarker(orURL)
			// feed data into the file-database (for updating)
			if err := saveStorage(s.Filename, s.AlURLStorage); err != nil {
				s.logger.Info("Failed to store the file for reading!", zap.String("filename", config.StartOptions.Filename), zap.Error(err)) // TODO mb I should panic here?
				return 1, err
			}
			// we have handled it for userID needed => break
			return 1, nil
		}
	}
	// new user init
	// We haven't found the user with userID in the storage => add new userID
	s.AlURLStorage = append(s.AlURLStorage, newStorageWithUserID(userID, alias, addExistMarker(orURL)))
	// feed data into the file-database (for updating)
	if err := saveStorage(s.Filename, s.AlURLStorage); err != nil {
		s.logger.Info("Failed to store the file for reading!", zap.String("filename", config.StartOptions.Filename), zap.Error(err)) // TODO mb I should panic here?
		return 1, err
	}
	return 1, nil

}

// GetDataByUserID gets short URLs and original URLs of a particular user.
func (s *Storage) GetDataByUserID(_ context.Context, userID string) (*[]models.UserURLS, error) {
	if s.AlURLStorage == nil {
		er := fmt.Errorf("no data for %v", userID)
		return nil, er
	}
	for _, userIDStorage := range s.AlURLStorage {
		if userIDStorage.UserID == userID {
			var dataOfuserID = make([]models.UserURLS, 0, 20)
			userIDStorage.mx.RLock()
			defer userIDStorage.mx.RUnlock()
			for al, url := range userIDStorage.AlURLStorage {
				dataOfuserID = append(dataOfuserID, models.UserURLS{ShortURL: al, OriginalURL: delMarker(url)})
			}
			return &dataOfuserID, nil

		}
	}
	// such a userID doesn't have any record
	er := fmt.Errorf("no data for %v", userID)
	return nil, er
}

// Ping is a blind plug to be able to implement the interface. It's used in pg package to ping a postgres database.
func (s *Storage) Ping(_ context.Context) error {
	return errors.New("DB isn't available")
}

// GetAlByURL is used just in pg option, HERE it's a blind plug.
func (s *Storage) GetAlByURL(_ context.Context, _ string) (string, error) { return "", nil }

// SaveBatchURL saves the batch of URLs in a file database.
func (s *Storage) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	const op = "file.SaveBatchURL"
	// create the arrays of structs for response
	toResp := make([]models.JSONBatStructToSerResp, 0) // TODO [MENTOR]: is it ok allocation? why len(*corURLSh) is false?
	// saving ..
	for _, idURLSh := range *corURLSh {
		_, err := s.StoreAlURL(ctx, idURLSh.ShortURL, idURLSh.OriginalURL, "") // TODO: how to use _ here?
		if err != nil {
			s.logger.Info(op, zap.String("filename", config.StartOptions.Filename), zap.Error(err))
			return nil, err
		}
		// molding the object for response (in controller)
		toResp = append(toResp, models.JSONBatStructToSerResp{
			CorrelationID: idURLSh.CorrelationID,
			ShortURL:      idURLSh.ShortURL,
		})
	}
	return &toResp, nil
}

// GetOrURLByAl returns #1st found# original url notwithstanding what user created it
func (s *Storage) GetOrURLByAl(_ context.Context, alias string) (string, error) {
	const op = "file.GetOrURLByAl"
	// Если горутина собирается читать данные, то она вызывает метод RLock(). Метод RLock() не
	//даёт начать запись пока не будут завершены все операции чтения.
	if s.AlURLStorage == nil { //TODO [MENTOR] IS IT NEEDED HERE???!
		er := fmt.Errorf("original URL for %v doesn't exist in the file-DB", alias)
		return "", er
	}
	for _, userIDStorage := range s.AlURLStorage {
		userIDStorage.mx.RLock() //// TODO: if I del it I don't have an error! #############################
		defer userIDStorage.mx.RUnlock()
		if orURL, ok := userIDStorage.AlURLStorage[alias]; ok {
			if isExist(orURL) { // CHECK the 1st founded URL !
				return delMarker(orURL), nil
			}
			// URL doesn't exist!
			return "", fmt.Errorf("%s: %w", op, merrors.ErrURLNotFound)
		}
	}
	er := fmt.Errorf("original URL for %v doesn't exist in the file-DB", alias)
	return "", er

}

// DeleteURLsByUserID deletes some records from the database by UserID (set flag is_deleted in true state)
// the func doesn't use transaction just to implement  a multi stream concept.
func (s *Storage) DeleteURLsByUserID(_ context.Context, userID string, aliasToDel []string) error {
	const op = "file.DeleteURLsByUserID"
	if s.AlURLStorage == nil {
		return fmt.Errorf("[INFO] file DB is empty")
	}
	for _, userIDStorage := range s.AlURLStorage {
		if userIDStorage.UserID == userID {
			userIDStorage.mx.Lock()
			defer userIDStorage.mx.Unlock()
			for _, alToDel := range aliasToDel {
				// if we have such alias in the memory => del this
				if _, ok := userIDStorage.AlURLStorage[alToDel]; ok {
					userIDStorage.AlURLStorage[alToDel] = changeExistToDelMarker(userIDStorage.AlURLStorage[alToDel])
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
	const op = "file.CheckIfUserExists"
	for _, userIDStorage := range s.AlURLStorage {
		if userIDStorage.UserID == userID {
			return nil
		}
	}
	return fmt.Errorf("[%s]: %w", op, merrors.ErrUserNotFound)
}

// FILE SAVING PART //

func getStorage(filename string) (*Storage, error) {
	opnFile, err := NewConsumer(filename)
	if err != nil {
		//s.logger.Error("Failed to open the file for reading!", zap.String("filename", filename), zap.Error(merrors))
		return nil, err
	}
	defer opnFile.Close()
	// read data nd get the link
	gotData, err := opnFile.ReadEvent()
	//fmt.Println("gotData = ", gotData)
	if err != nil {
		if errors.Is(err, io.EOF) {
			//s.logger.Info("File is empty (reading)", zap.String("filename", filename))
			// return empty storage
			//return &Storage{AlURLStorage: storageWithUserID{AlURLStorage: make(map[string]string)}}, nil
			return &Storage{}, nil
		}
		// handle other possible errors
		//s.logger.Error("Failed to read the file for reading!", zap.String("filename", filename), zap.Error(merrors))
		return nil, err
	}
	// add mutexes for structures which have been read from the file
	for _, storWithID := range gotData {
		storWithID.mx = &sync.RWMutex{}
	}
	return &Storage{AlURLStorage: gotData}, nil
}

// func (s *safeStorage) saveStorage(filename string) error {
func saveStorage(filename string, toSave []*storageWithUserID) error {
	opnFile, err := NewProducer(filename)
	if err != nil {
		//s.logger.Error("Failed to open the file for saving!", zap.String("filename", filename), zap.Error(merrors))
		return err
	}
	defer opnFile.Close()
	if err := opnFile.WriteEvent(toSave); err != nil {
		return err
	}
	return nil
}

// DeleteURLsByUserID is a blind plug here

// file-saving auxiliary code
type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

// NewProducer returns the structure with os.File and buffer elements to implement file database then (WRITER).
func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

// WriteEvent - flush data into the file database.
func (p *Producer) WriteEvent(event []*storageWithUserID) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

// Close closes os.File
func (p *Producer) Close() error {
	// закрываем файл
	return p.file.Close()
}

// Consumer - the consumer
type Consumer struct {
	file *os.File
	// добавляем reader в Consumer
	reader *bufio.Reader
}

// NewConsumer returns the structure with os.File and buffer elements to implement file database then (READER).
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый Reader
		reader: bufio.NewReader(file),
	}, nil
}

// ReadEvent returns unmarshalled data in the struct.
func (c *Consumer) ReadEvent() ([]*storageWithUserID, error) {
	// читаем данные до символа переноса строки
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	// преобразуем данные из JSON-представления в структуру
	//event := Storage{AlURLStorage: storageWithUserID{AlURLStorage: make(map[string]string)}}
	event := []*storageWithUserID{}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// Close closes os.File.
func (c *Consumer) Close() error {
	return c.file.Close()
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
