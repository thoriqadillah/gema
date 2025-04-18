package gema

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const LocalStorage StorageName = "local"

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

func init() {
	RegisterStorage(LocalStorage, createLocalStorage)
}
