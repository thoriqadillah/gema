package gema

import (
	"io"
	"os"
	"path/filepath"
)

const LocalStorage StorageName = "local"

type local struct {
	tmpDir string
}

func createLocal(option *StorageOption) Storage {
	return &local{
		tmpDir: option.TempDir,
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

	return filename, nil
}

func (l *local) Delete(filename string) error {
	return os.Remove(l.tmpDir + "/" + filename)
}

func init() {
	RegisterStorage(LocalStorage, createLocal)
}
