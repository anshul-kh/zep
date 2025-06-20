package helper

import (
	"fmt"
	"path/filepath"
)

func GetAbsPath(relPath string) (string, error) {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute file path:%w", err)
	}
	return absPath, nil
}
