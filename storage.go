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
	// Register will be used to register the storage implementation to the registry
	// and perform necessary operation along the way. Register must provide the storage implementation
	// privately if you want it to be injected and be reusable inside the Module. Otherwise return nil
	Register(registry StorageRegistry) fx.Option

	// Module will be used to provide additional dependencies
	// to the storage register and the whole app. Return nil if you don't want to provide anything
	Module() fx.Option
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

	fxOptions := make([]fx.Option, 0)
	for _, provider := range providers {
		storageFx := []fx.Option{}
		if registry := provider.Register(storageMap); registry != nil {
			storageFx = append(storageFx, registry)
		}

		if storageModule := provider.Module(); storageModule != nil {
			storageFx = append(storageFx, storageModule)
		}

		fxOptions = append(fxOptions, fx.Module("storage.provider", storageFx...))
	}

	fxOptions = append(fxOptions, fx.Provide(func() StorageFactory {
		return createStorage(storageMap)
	}))

	return fx.Module("storage", fxOptions...)
}
