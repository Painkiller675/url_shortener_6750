package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"go.uber.org/zap"
	"io"
	"os"
	"sync"
)

// JSONStruct is used to unmarshal js request nd send js response in CreateShortURLJSONHandler
type JSONStructSh struct {
	ShURL string `json:"result"`
}
type JSONStructOr struct {
	OrURL string `json:"url"`
}

type Storage struct {
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

// TODO constructor safeStorage
type safeStorage struct {
	AlURLStorage map[string]string `json:"al_url_storage"`
	mx           *sync.RWMutex     `json:"-"` // TODO pointer or not??
	logger       *zap.Logger       `json:"-"`
}

func newSafeStorage(logger *zap.Logger) *safeStorage {
	stor, err := getStorage(config.StartOptions.Filename)
	if err != nil {
		panic(err) // TODO HANDLE THAT! add err in signature??

	}
	return &safeStorage{
		AlURLStorage: stor.AlURLStorage, // mb save all the struct but wht about logger etc?
		mx:           &sync.RWMutex{},
		logger:       logger,
	}
}

// to send it to handler i should use func кот прин об или струтуру (applic1		 handler ..)

func (s *safeStorage) StoreAlURL(alias string, orURL string) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.AlURLStorage[alias] = orURL
	if err := saveStorage(config.StartOptions.Filename, s); err != nil {
		s.logger.Info("Failed to store the file for reading!", zap.String("filename", config.StartOptions.Filename), zap.Error(err)) // TODO mb I should panic here?
	}

}

func (s *safeStorage) GetOrURL(alias string) (string, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if orURL, ok := s.AlURLStorage[alias]; ok {
		return orURL, nil
	}
	return "", fmt.Errorf("original URL for %v doesn't exist in the DB", alias) //TODO handle that!
}

// READ/WRITE INTO THE FILE!
// TODO mb I should put it into safeStorage constructor to use logger?
// func (s *safeStorage) getStorage(filename string) (*safeStorage, error) {
func getStorage(filename string) (*safeStorage, error) {
	opnFile, err := NewConsumer(filename)
	if err != nil {
		//s.logger.Error("Failed to open the file for reading!", zap.String("filename", filename), zap.Error(err))
		return nil, err
	}
	defer opnFile.Close()
	// read data nd get the link
	gotData, err := opnFile.ReadEvent()
	if err != nil {
		if err == io.EOF {
			//s.logger.Info("File is empty (reading)", zap.String("filename", filename))
			// return empty storage
			return &safeStorage{
				AlURLStorage: make(map[string]string),
			}, nil
		}
		// handle other possible errors
		//s.logger.Error("Failed to read the file for reading!", zap.String("filename", filename), zap.Error(err))
		return nil, err
	}
	return gotData, nil
}

// func (s *safeStorage) saveStorage(filename string) error {
func saveStorage(filename string, toSave *safeStorage) error {
	opnFile, err := NewProducer(filename)
	if err != nil {
		//s.logger.Error("Failed to open the file for saving!", zap.String("filename", filename), zap.Error(err))
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

func (p *Producer) WriteEvent(event *safeStorage) error {
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
func (c *Consumer) ReadEvent() (*safeStorage, error) {
	// читаем данные до символа переноса строки
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	fmt.Println(string(data))
	// преобразуем данные из JSON-представления в структуру
	event := safeStorage{}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
