package storage

type Storage interface {
	Apply(msg string) bool
}