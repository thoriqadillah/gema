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
)

const LocalStorage StorageName = "local"

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

type localStorage struct {
	tmpDir        string
	fullRoutePath string
}

func createLocalStorage(option *StorageOption) Storage {
	return &localStorage{
		tmpDir:        option.TempDir,
		fullRoutePath: option.FullRoutePath,
	}
}

func (l *localStorage) Serve(filename string) (io.ReadCloser, error) {
	return os.Open(l.tmpDir + "/" + filename)
}

func (l *localStorage) Upload(filename string, src io.Reader) (string, error) {
	file := filepath.Join(l.tmpDir, filename)

	if err := os.MkdirAll(l.tmpDir, 0755); err != nil {
		return "", err
	}

	dst, err := os.Create(file)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("%s/%s", l.fullRoutePath, filename)
	return path, nil
}

func (l *localStorage) Delete(filename string) error {
	return os.Remove(l.tmpDir + "/" + filename)
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

func init() {
	RegisterStorage(LocalStorage, createLocalStorage)
}
