package fileglob

import (
	"os"
	"log"
	"runtime"
	"path/filepath"
	"github.com/intdxdt/extfilter"
)

const GlobStreamLimit = 100

//filter
func filter(results *[]string, filters []string, strict bool, ignoreDirs []string, flush ...func(bool)) filepath.WalkFunc {
	extFilter := extfilter.NewExtensionFilters(filters, strict)
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dir := filepath.Base(path)
			for _, d := range ignoreDirs {
				if d == dir {
					return filepath.SkipDir
				}
			}
		}
		if !info.IsDir() && info.Mode().IsRegular() {
			if extFilter.Match(path) {
				*results = append(*results, path)
				for _, fn := range flush {
					fn(false)
				}
			}
		}
		return nil
	}
}

//Glob
func Glob(directory string, filters []string, strict bool, ignoreDirs []string) ([]string, error) {
	results := make([]string, 0)
	err := filepath.Walk(directory, filter(&results, filters, strict, ignoreDirs))
	return results, err
}

//Glob
func GlobStream(directory string, filters []string, strict bool,
	ignoreDirs []string, bufferSize ...int) <-chan []string {
	size := GlobStreamLimit
	if len(bufferSize) > 0 {
		size = bufferSize[0]
	}
	stream := make(chan []string)
	buffer := make([]string, 0, size)

	flush := func(drain bool) {
		if drain || len(buffer) >= size {
			stream <- buffer
			buffer = make([]string, 0, size)
		}
	}

	go func() {
		defer close(stream)
		err := filepath.Walk(directory, filter(&buffer, filters, strict, ignoreDirs, flush))
		if err != nil {
			log.Fatalln(err)
		}
		if len(buffer) > 0 {
			flush(true)
		}
		runtime.Gosched()
	}()

	return stream
}
