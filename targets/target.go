package targets

type Target interface {
	Upload(filePath string) error
	Clean(backupExpirationDays int) error
}
