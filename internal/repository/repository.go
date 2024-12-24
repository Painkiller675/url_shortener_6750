package repository

import (
	"fmt"
	"sync"
)

// JSONStruct is used to unmarshal js request nd send js response in CreateShortURLJSONHandler
type JSONStructSh struct {
	ShURL string `json:"result"`
}
type JSONStructOr struct {
	OrURL string `json:"url"`
}

// TODO constructor safeStruct
type safeStruct struct {
	AlURLStorage map[string]string `json:"al_url_storage"`
	mx           sync.RWMutex      `json:"-"`
}

// to send it to handler i should use func кот прин об или струтуру (applic1		 handler ..)

func (s *safeStruct) StoreAlURL(alias string, orURL string) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.AlURLStorage[alias] = orURL
}

func (s *safeStruct) getOrURL(alias string) (string, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if orURL, ok := s.AlURLStorage[alias]; ok {
		return orURL, nil
	}
	return "", fmt.Errorf("original URL for %v doesn't exist in the DB", alias)
}

func newSafeStruct() *safeStruct {
	var s safeStruct
	s.AlURLStorage = make(map[string]string)
	return &s
}

// TODO public constructor return SafeStruct and in struct that it returns thre will be get nd write
var orAlURLStorage = newSafeStruct() // ALIAS - orURL

func WriteURL(newAl string, newOrURL string) {
	// check if such url already exists if exists => change that
	orAlURLStorage.StoreAlURL(newAl, newOrURL)
	fmt.Println("from write", orAlURLStorage.AlURLStorage)
}

func GetShortURL(alias string) (string, error) {
	curAl, err := orAlURLStorage.getOrURL(alias)
	if err != nil {
		return "", err
	}

	return curAl, nil
}

/*
// Open / close file
type Consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
}

// NewConsumer to open file
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый scanner
		scanner: bufio.NewScanner(file),
	}, nil
}

// Read to read from file
func (c *Consumer) ReadEvent() (*Event, error) {
	// одиночное сканирование до следующей строки
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	event := Event{}
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
*/
