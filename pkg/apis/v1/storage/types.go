package storage

// Storage interface of all storage plugin
// includes: memory,fileSystem,net fileSystem will implement these
type Storage interface {
	Init(opt ...Option) (bool, error)
}

// Option operate to init storage
type Option func(Storage)
