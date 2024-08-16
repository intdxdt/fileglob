package fileglob

import (
	"github.com/intdxdt/extfilter"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const globStreamLimit = 64

func Glob(directory string, filters []string, strict bool, ignoreDirs []string) ([]string, error) {
	realPath, err := getRealPath(directory)
	if err != nil {
		return nil, err
	}

	var results = make([]string, 0, 32)
	ignoreDirSet := make(map[string]struct{}, len(ignoreDirs))
	for _, dir := range ignoreDirs {
		ignoreDirSet[dir] = struct{}{}
	}

	extFilter := extfilter.NewExtensionFilters(filters, strict)
	err = filepath.Walk(realPath, filter(&results, extFilter, ignoreDirSet))
	return results, err
}

func GlobStream(directory string, filters []string, strict bool, ignoreDirs []string, bufferSize ...int) <-chan []string {
	size := globStreamLimit
	if len(bufferSize) > 0 {
		size = bufferSize[0]
	}
	stream := make(chan []string, 1)
	buffer := make([]string, 0, size)

	flush := func(drain bool) {
		if drain || len(buffer) >= size {
			stream <- buffer
			buffer = make([]string, 0, size)
		}
	}

	go func() {
		defer close(stream)
		realPath, err := getRealPath(directory)
		if err != nil {
			log.Fatalln(err)
		}

		ignoreDirSet := make(map[string]struct{}, len(ignoreDirs))
		for _, dir := range ignoreDirs {
			ignoreDirSet[dir] = struct{}{}
		}

		extFilter := extfilter.NewExtensionFilters(filters, strict)
		err = filepath.Walk(realPath, filter(&buffer, extFilter, ignoreDirSet, flush))
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

func filter(results *[]string, extFilter *extfilter.ExtensionFilter, ignoreDirs map[string]struct{}, flush ...func(bool)) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, ignored := ignoreDirs[filepath.Base(path)]; ignored {
				return filepath.SkipDir
			}
		} else if info.Mode().IsRegular() && extFilter.Match(path) {
			*results = append(*results, path)
			for _, fn := range flush {
				fn(false)
			}
		}
		return nil
	}
}

func getRealPath(dir string) (string, error) {
	stat, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}

	if stat.Mode()&os.ModeSymlink != 0 {
		resolvedDir, err := os.Readlink(dir)
		if err != nil {
			return "", err
		}
		return resolvedDir, nil
	}

	return dir, nil
}
