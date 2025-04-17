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
}

type StorageFactory func(option *StorageOption) Storage

var providers = map[StorageName]StorageFactory{}

type StorageOptionFunc func(*StorageOption)

func WithTempDir(tempDir string) StorageOptionFunc {
	return func(o *StorageOption) {
		o.TempDir = tempDir
	}
}

func StorageModule(name StorageName, opts ...StorageOptionFunc) fx.Option {
	return fx.Module("storage", fx.Provide(
		func() StorageFacade {
			return NewStorage(name, opts...)
		},
	))
}

func NewStorage(name StorageName, opts ...StorageOptionFunc) StorageFacade {
	pwd, _ := os.Getwd()
	opt := &StorageOption{
		TempDir: pwd + "/storage/tmp",
	}

	for _, option := range opts {
		option(opt)
	}

	provider, ok := providers[name]
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
	return NewStorage(driver, opts...)
}

func RegisterStorage(name StorageName, impl StorageFactory) {
	providers[name] = impl
}
