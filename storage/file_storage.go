package storage

import (
	"io"
	"os"
	"sync"
)

type FileStorage struct {
	Lock    sync.Mutex
	LogFile *os.File
}

func NewFileStorage(filename string) Storage {
	f, _ := os.Create(filename)
	return &FileStorage{
		Lock:    sync.Mutex{},
		LogFile: f,
	}
}

func (s *FileStorage) Apply(msg string) (res bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	_, err := io.WriteString(s.LogFile, msg)
	if err != nil {
		panic(err)
	}

	return true
}
