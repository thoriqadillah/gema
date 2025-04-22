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

const LocalStorageName StorageName = "local"

type LocalStorageOption struct {
	TempDir       string
	FullRoutePath string
}

type localStorage struct {
	opt *LocalStorageOption
}

func newLocalStorage(opt *LocalStorageOption) *localStorage {
	return &localStorage{
		opt: opt,
	}
}

func (l *localStorage) Serve(filename string) (io.ReadCloser, error) {
	return os.Open(l.opt.TempDir + "/" + filename)
}

func (l *localStorage) Upload(filename string, src io.Reader) (string, error) {
	file := filepath.Join(l.opt.TempDir, filename)

	if err := os.MkdirAll(l.opt.TempDir, 0755); err != nil {
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

	path := fmt.Sprintf("%s/%s", l.opt.FullRoutePath, filename)
	return path, nil
}

func (l *localStorage) Delete(filename string) error {
	return os.Remove(l.opt.TempDir + "/" + filename)
}

type storageController struct {
	storage   Storage
	routePath string
}

func newStorageController(opt *LocalStorageOption, storage Storage) Controller {
	url, err := url.Parse(opt.FullRoutePath)
	if err != nil {
		log.Fatalf("[Gema] Invalid route path %s", opt.FullRoutePath)
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

type localStorageProvider struct {
	opt *LocalStorageOption
}

func LocalStorageProvider(opt *LocalStorageOption) StorageProvider {
	return &localStorageProvider{opt}
}

func (l *localStorageProvider) Register(registry StorageRegistry) fx.Option {
	return fx.Provide(fx.Private, func() Storage {
		storage := newLocalStorage(l.opt)
		registry.Register(LocalStorageName, storage)

		return storage
	})
}

func (l *localStorageProvider) provideOption() *LocalStorageOption {
	return l.opt
}

func (l *localStorageProvider) Module() fx.Option {
	return fx.Module("storage.controller",
		fx.Provide(fx.Private, l.provideOption),
		RegisterController(newStorageController),
	)
}
