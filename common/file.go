package common

import (
	"os"
	"path/filepath"
)

var CurrentDirectory string

func init() {
	if cwd, err := os.Getwd(); err == nil {
		if directory, err := filepath.Abs(cwd); err == nil {
			CurrentDirectory = directory
		}
	}
}
