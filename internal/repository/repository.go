package repository

import (
	"fmt"
	"sync"
)

type safeStruct struct {
	AlURLStorage map[string]string
	mx           sync.RWMutex
}

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
