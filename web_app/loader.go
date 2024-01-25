package web_app

import (
	"errors"
	"io/fs"
	"sync"
)

var ErrInvalidFilePath error = errors.New("invalid file path")

// it is not thread-safe
type store struct {
	kv_store map[string]string
}

func NewStore() *store {
	return &store{
		kv_store: make(map[string]string),
	}
}

var partialsManagerInstance *partialsManager
var lock *sync.Mutex = &sync.Mutex{}


// I don't know if this is a good solution or not
// TODO: consider this later.
// I'm using this because you can't compare a PartialsManger instance
var isInstanceCreated bool = false

// The PartialsManger stores the path for each partial, prevents duplicates and is a singleton to keep things organized
// NOT thread-safe
type partialsManager struct {
	fileSystem fs.FS
	store      *store
}

func (s *store) get(key string) (value string, err error) {
	value, ok := s.kv_store[key]

	if !ok {
		return "", errors.New("invalid key")
	}

	return value, nil
}

func (s *store) put(key string, value string) (string, error) {
	_, ok := s.kv_store[key]
	if !ok {
		s.kv_store[key] = value
		return value, nil
	}

	return value, errors.New("attempted to enter duplicate key")
}

func ClearPartialsManager() {
	lock.Lock()
	defer lock.Unlock()

	if isInstanceCreated == true {
		partialsManagerInstance.Clear()
		isInstanceCreated = false
	}
}

func GetPartialsManager(filesystem fs.FS) *partialsManager {
	lock.Lock()
	defer lock.Unlock()
	if isInstanceCreated == false {
		instance := &partialsManager{
			fileSystem: filesystem,
			store:      NewStore(),
		}
		partialsManagerInstance = instance
		isInstanceCreated = true
		return instance
	}
	return partialsManagerInstance
}

func (p *partialsManager) RegisterPartial(partialName, fileName string) error {
	// check that the file is valid
	_, err := fs.ReadFile(p.fileSystem, fileName)
	if err != nil {
		return ErrInvalidFilePath
	}

	_, err = p.store.put(partialName, fileName)

	if err != nil {
		return err
	}

	return nil
}

func (p *partialsManager) GetPartial(partialName string) (partialPath string, error error) {
	value, err := p.store.get(partialName)
	if err != nil {
		return "", err
	}
	return value, nil
}

// Clear returns the partialsManager to an empty state
func (p *partialsManager) Clear() error{
	newStore := NewStore()
	p.store = newStore
	// if there's ever a need for an error while clearing.
	return nil
}
