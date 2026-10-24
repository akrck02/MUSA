package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheMemoryProfileIsWrittenAndSaysHowMuchIsGos(test *testing.T) {
	path := filepath.Join(test.TempDir(), "memory.pprof")
	var report strings.Builder
	if err := writeMemoryProfile(path, &report); err != nil {
		test.Fatal(err)
	}
	if information, err := os.Stat(path); err != nil || information.Size() == 0 {
		test.Errorf("no profile written: %v", err)
	}
	if !strings.Contains(report.String(), "MB in use") {
		test.Errorf("the report says %q", report.String())
	}
}
