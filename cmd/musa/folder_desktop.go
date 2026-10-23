//go:build !android

package main

import (
	"os"
	"path/filepath"
)

// dataFolder is MUSA's folder in the user's configuration folder.
func dataFolder() (string, error) {
	folder, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(folder, "musa"), nil
}
