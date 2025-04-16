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

type StorageFacade struct {
	Storage
}

type Option struct {
	// The directory to store temporary files for local storage
	tempDir string
}

type Factory func(option *Option) Storage

var providers = map[StorageName]Factory{}

type OptionFunc func(*Option)

func WithTempDir(tempDir string) OptionFunc {
	return func(o *Option) {
		o.tempDir = tempDir
	}
}

type StorageOption struct {
	name StorageName
	opts []OptionFunc
}

func StorageModule(name StorageName, opts ...OptionFunc) fx.Option {
	return fx.Module("storage", fx.Provide(
		func() *StorageFacade {
			return New(name, opts...)
		},
	))
}

func New(name StorageName, opts ...OptionFunc) *StorageFacade {
	pwd, _ := os.Getwd()
	opt := &Option{
		tempDir: pwd + "/storage/tmp",
	}

	for _, option := range opts {
		option(opt)
	}

	provider, ok := providers[name]
	if !ok {
		log.Fatalf("Storage with %s provider not found", name)
		return nil
	}

	storage := provider(opt)
	return withFacade(storage)
}

func withFacade(s Storage) *StorageFacade {
	return &StorageFacade{
		Storage: s,
	}
}

func (s *StorageFacade) Use(driver StorageName, opts ...OptionFunc) Storage {
	return New(driver, opts...)
}

func Register(name StorageName, impl Factory) {
	providers[name] = impl
}
