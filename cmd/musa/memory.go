package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
)

// bytesPerMegabyte turns the runtime's byte counts into the megabytes a person reads.
const bytesPerMegabyte = 1 << 20

// writeMemoryProfile saves what Go holds when the window closes, for go tool pprof, and tells on
// standard error how much of the process's memory is Go's: the rest belongs to the system's
// libraries and the graphics driver.
func writeMemoryProfile(path string, report io.Writer) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	runtime.GC() // the profile shows what is still held, not what is waiting to be collected
	if err := pprof.WriteHeapProfile(file); err != nil {
		return err
	}
	var statistics runtime.MemStats
	runtime.ReadMemStats(&statistics)
	fmt.Fprintf(report, "musa: Go holds %d MB in use, %d MB taken from the system; profile in %s\n",
		statistics.HeapInuse/bytesPerMegabyte, statistics.Sys/bytesPerMegabyte, path)
	return nil
}
