package gema

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const LocalStorage StorageName = "local"

type local struct {
	tmpDir        string
	fullRoutePath string
}

func createLocal(option *StorageOption) Storage {
	return &local{
		tmpDir:        option.TempDir,
		fullRoutePath: option.FullRoutePath,
	}
}

func (l *local) Serve(filename string) (io.ReadCloser, error) {
	return os.Open(l.tmpDir + "/" + filename)
}

func (l *local) Upload(filename string, src io.Reader) (string, error) {
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

func (l *local) Delete(filename string) error {
	return os.Remove(l.tmpDir + "/" + filename)
}

func init() {
	RegisterStorage(LocalStorage, createLocal)
}
