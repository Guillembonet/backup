package sources

import "fmt"

type Source interface {
	Backup(destination string) error
}

var (
	ErrNothingToBackup = fmt.Errorf("nothing to backup")
)
