package gema

import (
	"io"
	"log"

	"go.uber.org/fx"
)

type StorageName string

type Storage interface {
	Serve(filename string) (io.ReadCloser, error)

	// Upload will upload a file to the storage and return the url of the file
	Upload(filename string, src io.Reader) (string, error)
	Delete(filename string) error
}

type StorageFactory interface {
	// Disk will return the storage implementation
	Disk(driver StorageName) Storage
}

type StorageRegistry map[StorageName]Storage

func (s StorageRegistry) Register(name StorageName, storage Storage) {
	s[name] = storage
}

type StorageProvider interface {
	// Register will be used to register your storage implementation and returns your storage module.
	// In the register, you will be provided with the storage registry to register your storage implementation.
	// Register must return your storage module with fx.Option. But remember to make your storage
	// implementation private. Otherwise, it will collide with other storage implementations
	Register() fx.Option
}

type storageFactory struct {
	registry StorageRegistry
}

func createStorage(s StorageRegistry) StorageFactory {
	return &storageFactory{s}
}

func (s *storageFactory) Disk(driver StorageName) Storage {
	storage, ok := s.registry[driver]
	if !ok {
		log.Fatalf("[Gema] Storage with %s provider not found", driver)
		return nil
	}

	return storage
}

// StorageModule is a module that will register the storage provider
// and provide the storage factory to the app.
// The storage provider must implement the StorageProvider interface
func StorageModule(providers ...StorageProvider) fx.Option {
	storageMap := StorageRegistry{}

	fxOptions := []fx.Option{
		fx.Provide(fx.Private, func() StorageRegistry {
			return storageMap
		}),
	}

	for _, provider := range providers {
		fxOptions = append(fxOptions, provider.Register())
	}

	fxOptions = append(fxOptions, fx.Provide(createStorage))

	return fx.Module("storage", fxOptions...)
}
