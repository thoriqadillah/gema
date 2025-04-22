package gema

import (
	"io"
	"log"
	"os"

	"go.uber.org/fx"
)

type StorageName string

type Storage interface {
	Serve(filename string) (io.ReadCloser, error)

	// Upload will upload a file to the storage and return the url of the file
	Upload(filename string, src io.Reader) (string, error)
	Delete(filename string) error
}

type StorageFacade interface {
	Storage
	Use(driver StorageName, opts ...StorageOptionFunc) Storage
}

type StorageOption struct {
	// The directory to store temporary files for local storage.
	// Default is `pwd + "/storage/tmp"`
	TempDir string

	// The route path to serve the file removetly
	// Default is `/storage/:filename`
	FullRoutePath string
}

type StorageFactory func(option *StorageOption) Storage

var storageProviders = map[StorageName]StorageFactory{}

type StorageOptionFunc func(*StorageOption)

// StorageModule is a module to provide storage service with its controller to serve local storage
func StorageModule(name StorageName, opts ...StorageOptionFunc) fx.Option {
	pwd, _ := os.Getwd()
	opt := &StorageOption{
		TempDir: pwd + "/storage/tmp",
	}

	for _, option := range opts {
		option(opt)
	}

	return fx.Module("storage",
		fx.Provide(func() StorageFacade {
			return newStorage(name, opt)
		}),
		fx.Provide(fx.Private, func() *StorageOption {
			return opt
		}),
		RegisterController(newStorageController),
	)
}

func newStorage(name StorageName, opt *StorageOption) StorageFacade {
	provider, ok := storageProviders[name]
	if !ok {
		log.Fatalf("[Gema] Storage with %s provider not found", name)
		return nil
	}

	storage := provider(opt)
	return withFacade(storage)
}

type storageFacade struct {
	Storage
}

func withFacade(s Storage) StorageFacade {
	return &storageFacade{s}
}

func (s *storageFacade) Use(driver StorageName, opts ...StorageOptionFunc) Storage {
	pwd, _ := os.Getwd()
	opt := &StorageOption{
		TempDir: pwd + "/storage/tmp",
	}

	for _, option := range opts {
		option(opt)
	}

	return newStorage(driver, opt)
}

func RegisterStorage(name StorageName, impl StorageFactory) {
	storageProviders[name] = impl
}
