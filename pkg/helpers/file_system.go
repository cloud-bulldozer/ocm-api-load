package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
)

// CreateFolder creates folder in the system
func CreateFolder(path string, logger logging.Logger) error {
	logger.Info(context.Background(), "Creating '%s' directory", path)
	folder, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(folder, os.FileMode(0755))
	if err != nil {
		return err
	}
	return nil
}

// CreateFile creates the file with the given name
func CreateFile(name, path string) (*os.File, error) {
	resultPath := filepath.Join(path, name)
	out, err := os.Create(resultPath)
	if err != nil {
		// Silently ignore pre-existing file.
		if err == os.ErrExist {
			return out, nil
		}
		return nil, fmt.Errorf("writing result: %v", err)
	}
	return out, nil
}
