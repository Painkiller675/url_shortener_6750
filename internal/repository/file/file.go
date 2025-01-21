package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"go.uber.org/zap"
	"io"
	"os"
	"sync"
)

type Storage struct {
	AlURLStorage map[string]string `json:"url_storage"`
	Filename     string            `json:"-"`
	mx           *sync.RWMutex     `json:"-"` // TODO pointer or not??
	Logger       *zap.Logger       `json:"-"` // TODO [MENTOR] make it public or private and why???
}

func NewStorage(filename string, logger *zap.Logger) *Storage {
	fmt.Println("[INFO] file storage is available")
	stor, err := getStorage(filename)
	if err != nil {
		logger.Fatal("[FATAL] file storage is not available", zap.Error(err))
	}
	return &Storage{
		AlURLStorage: stor.AlURLStorage, // mb save all the struct but wht about logger etc?
		mx:           &sync.RWMutex{},
		Logger:       logger,
		Filename:     filename,
	}
}

func (s *Storage) StoreAlURL(_ context.Context, alias string, orURL string) (int64, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.AlURLStorage[alias] = orURL
	if err := saveStorage(s.Filename, s); err != nil {
		s.Logger.Info("Failed to store the file for reading!", zap.String("filename", config.StartOptions.Filename), zap.Error(err)) // TODO mb I should panic here?
		return 1, err                                                                                                                // 1 - a blind plug
	}
	return 1, nil // a blind plug
}

// a blind plug to be able to implement the interface
func (s *Storage) Ping(ctx context.Context) error {
	fmt.Println("[INFO] ping from the file")
	return errors.New("DB isn't available")
}

// used just in pg option
func (s *Storage) GetAlByURL(_ context.Context, _ string) (string, error) { return "", nil }

func (s *Storage) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	const op = "file.SaveBatchURL"
	// create the arrays of structs for response
	toResp := make([]models.JSONBatStructToSerResp, 0) // TODO [MENTOR]: is it ok allocation? why len(*corURLSh) is false?
	// saving ..
	for _, idURLSh := range *corURLSh {
		_, err := s.StoreAlURL(ctx, idURLSh.ShortURL, idURLSh.OriginalURL) // TODO: how to use _ here?
		if err != nil {
			s.Logger.Info(op, zap.String("filename", config.StartOptions.Filename), zap.Error(err))
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
func (s *Storage) GetOrURLByAl(_ context.Context, alias string) (string, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if orURL, ok := s.AlURLStorage[alias]; ok {
		return orURL, nil
	}
	return "", fmt.Errorf("original URL for %v doesn't exist in the DB", alias) //TODO handle that!
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
	fmt.Println("gotData = ", gotData)
	if err != nil {
		if errors.Is(err, io.EOF) {
			//s.logger.Info("File is empty (reading)", zap.String("filename", filename))
			// return empty storage
			return &Storage{
				AlURLStorage: make(map[string]string),
			}, nil
		}
		// handle other possible errors
		//s.logger.Error("Failed to read the file for reading!", zap.String("filename", filename), zap.Error(merrors))
		return nil, err
	}
	return gotData, nil
}

// func (s *safeStorage) saveStorage(filename string) error {
func saveStorage(filename string, toSave *Storage) error {
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

// file-saving auxiliary code
type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

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

func (p *Producer) WriteEvent(event *Storage) error {
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

func (p *Producer) Close() error {
	// закрываем файл
	return p.file.Close()
}

type Consumer struct {
	file *os.File
	// добавляем reader в Consumer
	reader *bufio.Reader
}

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

// ReadEvent returns unmarshalled data in the struct
func (c *Consumer) ReadEvent() (*Storage, error) {
	// читаем данные до символа переноса строки
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	//fmt.Println(string(data))
	// преобразуем данные из JSON-представления в структуру
	event := Storage{AlURLStorage: make(map[string]string)}
	err = json.Unmarshal(data, &event)
	fmt.Println("AFTER unmarshall: ", event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
