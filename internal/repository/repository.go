package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
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

func newSafeStruct(fileData *safeStruct) *safeStruct {
	//var s safeStruct
	//s.AlURLStorage = make(map[string]string)
	//return &s
	var s safeStruct
	s.AlURLStorage = fileData.AlURLStorage
	return &s
}

// TODO public constructor return SafeStruct and in struct that it returns thre will be get nd write
var orAlURLStorage = newSafeStruct(InitStorage("./stor.json")) // ALIAS - orURL

func WriteURL(newAl string, newOrURL string) {
	// check if such url already exists if exists => change that
	orAlURLStorage.StoreAlURL(newAl, newOrURL)
	fmt.Println("from write", orAlURLStorage.AlURLStorage)
	SaveStorage(config.StartOptions.Filename, orAlURLStorage)
}

func GetShortURL(alias string) (string, error) {
	curAl, err := orAlURLStorage.getOrURL(alias)
	if err != nil {
		return "", err
	}

	return curAl, nil
}

func InitStorage(filename string) *safeStruct {
	cons, err := NewConsumer(filename)
	if err != nil {
		panic(err)
	}
	defer cons.Close()
	gotData, err := cons.ReadEvent()
	if err != nil {
		if err == io.EOF {
			var s safeStruct
			return &s
		}
		panic(err) // TODO how to handle it?
	}
	return gotData
}

func SaveStorage(filename string, structToWrite *safeStruct) {
	prod, err := NewProducer("./dir/json.txt")
	if err != nil {
		panic(err)
	}
	defer prod.Close()

	prod.WriteEvent(structToWrite)
}

// file-saving auxiliary code
type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteEvent(event *safeStruct) error {
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

func (c *Consumer) ReadEvent() (*safeStruct, error) {
	// читаем данные до символа переноса строки
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	fmt.Println(string(data))
	// преобразуем данные из JSON-представления в структуру
	event := safeStruct{}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
