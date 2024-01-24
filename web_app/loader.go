package web_app

import (
	"errors"
	"io/fs"
)

type store struct {
	kv_store map[string]string
}

func NewStore() *store {
	return &store{
		kv_store: make(map[string]string),
	}
}

var partialsManagerInstance *PartialsManager

// I don't know if this is a good solution or not
// TODO: consider this later.
// I'm using this because you can't compare a PartialsManger instance
var isInstanceCreated bool = false

// The PartialsManger stores the path for each partial, prevents duplicates and is a singleton to keep things organized
type PartialsManager struct {
	fileSystem fs.FS
	store      store
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

func GetPartialsManager(filesystem fs.FS) *PartialsManager {
	if isInstanceCreated == false {
		isInstanceCreated = true
		instance := &PartialsManager{
			fileSystem: filesystem,
			store:      *NewStore(),
		}
		partialsManagerInstance = instance
		return instance
	}
	return partialsManagerInstance
}

func (p *PartialsManager) RegisterPartial(partialName, fileName string) error {
	// check that the file is valid
	_, err := p.fileSystem.Open(fileName)
	if err != nil {
		return err
	}

	_, err = p.store.put(partialName, fileName)

	if err != nil {
		return err
	}

	return nil
}

func (p *PartialsManager) GetPartial(partialName string) (partialPath string, error error) {
	value, err := p.store.get(partialName)
	if err != nil {
		return "", err
	}
	return value, nil
}
