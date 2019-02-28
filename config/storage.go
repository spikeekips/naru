package config

import (
	"fmt"
	"os"
	"path/filepath"

	sebakstorage "boscoin.io/sebak/lib/storage"
	"github.com/spikeekips/cvc"

	"github.com/spikeekips/naru/common"
)

type Storage struct {
	cvc.BaseGroup
	Path string `flag-help:"storage path"`
}

func NewStorage() *Storage {
	return &Storage{
		Path: filepath.Join(common.CurrentDirectory, common.DefaultStoragePath),
	}
}

func (s Storage) ParsePath(i string) (string, error) {
	path := filepath.Join(common.CurrentDirectory, i)
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return "", fmt.Errorf("storage path is not directory")
		}
	}
	return path, nil
}

func (s *Storage) StorageConfig() *sebakstorage.Config {
	c, _ := sebakstorage.NewConfigFromString("file://" + s.Path)

	return c
}

func (s *Storage) FlagValuePath() string {
	n, err := filepath.Rel(common.CurrentDirectory, s.Path)
	if err != nil {
		return s.Path
	}
	return n
}
