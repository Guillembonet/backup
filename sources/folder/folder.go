package folder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Source struct {
	source string
}

func NewSource(source string) (*Source, error) {
	return &Source{
		source: strings.TrimSuffix(source, "/"),
	}, nil
}

func (s *Source) Backup(destination string) error {
	// get the base name of the source directory
	sourceBase := filepath.Base(s.source)
	// create the destination directory if it does not exist
	destination = filepath.Join(destination, sourceBase)
	err := recursiveCopy(s.source, destination)
	if err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}

func recursiveCopy(source string, destination string) error {
	// create the destination directory if it does not exist
	err := os.MkdirAll(destination, os.ModePerm)
	if err != nil {
		return err
	}

	// get a list of files in the source directory
	files, err := filepath.Glob(filepath.Join(source, "*"))
	if err != nil {
		return err
	}

	// copy each file in the source directory to the destination directory
	for _, file := range files {
		if isDir(file) {
			// recursively copy the directory
			dirName := filepath.Base(file)
			err = recursiveCopy(file, filepath.Join(destination, dirName))
			if err != nil {
				return err
			}
		} else {
			// copy the file
			dest := filepath.Join(destination, filepath.Base(file))
			src, err := os.Open(file)
			if err != nil {
				return err
			}
			defer src.Close()

			dst, err := os.Create(dest)
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
