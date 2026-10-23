//go:build android

package main

import "os"

// dataFolder is the application's own folder: Go Renderer points the temporary folder there, and
// Android removes it with the application.
func dataFolder() (string, error) { return os.TempDir(), nil }
