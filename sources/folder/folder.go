package folder

import (
	"io"
	"os"
	"path/filepath"

	"github.com/guillembonet/backup/sources"
)

type Source struct {
	source string
}

func NewSource(source string) (*Source, error) {
	return &Source{
		source: source,
	}, nil
}

func (s *Source) Backup(destination string) error {
	// get the base name of the source directory
	sourceBase := filepath.Base(s.source)
	// create the destination directory if it does not exist
	destination = filepath.Join(destination, sourceBase)
	err := os.MkdirAll(destination, 0755)
	if err != nil {
		return err
	}

	// get a list of files in the source directory
	files, err := filepath.Glob(filepath.Join(s.source, "*"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return sources.ErrNothingToBackup
	}

	// copy each file to the destination directory
	for _, file := range files {
		// create a new file in the destination directory with the same name as the source file
		dest := filepath.Join(destination, filepath.Base(file))
		src, err := os.Open(file)
		if err != nil {
			return err
		}
		defer src.Close()

		// copy the contents of the source file to the destination file
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
	return nil
}
