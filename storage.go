package gema

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
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

// WithStorageTempDir sets the temporary directory to store files.
// Default is `pwd + "/storage/tmp"`
func WithStorageTempDir(tempDir string) StorageOptionFunc {
	return func(o *StorageOption) {
		o.TempDir = tempDir
	}
}

// WithRoutePath sets the route path to serve the file remotely. Please provide the full path.
// Example: http://localhost:8000/storage.
// Required for local storage.
func WithStorageUrlPath(routePath string) StorageOptionFunc {
	return func(o *StorageOption) {
		o.FullRoutePath = routePath
	}
}

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
			fmt.Println("[Gema] Registering storage module")
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

type storageController struct {
	storage   StorageFacade
	routePath string
}

func newStorageController(option *StorageOption, storage StorageFacade) Controller {
	url, err := url.Parse(option.FullRoutePath)
	if err != nil {
		log.Fatalf("[Gema] Invalid route path %s", option.FullRoutePath)
	}

	return &storageController{
		storage:   storage,
		routePath: url.Path,
	}
}

func (s *storageController) serve(c echo.Context) error {
	filename := c.Param("filename")
	ext := filepath.Ext(filename)
	mimetype := mime.TypeByExtension(ext)

	file, err := s.storage.Serve(filename)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("File %s not found", filename))
	}
	defer file.Close()

	return c.Stream(http.StatusOK, mimetype, file)
}

func (s *storageController) CreateRoutes(r *echo.Group) {
	path := fmt.Sprintf("%s/:filename", s.routePath)
	r.GET(path, s.serve)
}
