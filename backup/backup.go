package backup

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/guillembonet/backup/config"
	"github.com/guillembonet/backup/sources"
	"github.com/guillembonet/backup/sources/folder"
	"github.com/guillembonet/backup/targets"
	"github.com/guillembonet/backup/targets/mega"
	"github.com/rs/zerolog/log"
	"github.com/xdg-go/pbkdf2"
)

type Backup struct {
	cfg     config.Backup
	sources []sources.Source
	targets []targets.Target
}

func New(cfg config.Backup) (*Backup, error) {
	sources := make([]sources.Source, len(cfg.Sources))
	for i, source := range cfg.Sources {
		switch source.Type {
		case "folder":
			s, err := folder.NewSource(source.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to create folder source: %w", err)
			}
			sources[i] = s
		default:
			return nil, fmt.Errorf("unknown source type: %s", source.Type)
		}
	}

	targets := make([]targets.Target, len(cfg.Targets))
	for i, target := range cfg.Targets {
		switch target.Type {
		case "mega":
			t, err := mega.NewClient(target.Config)
			if err != nil {
				return nil, fmt.Errorf("failed to create mega target: %w", err)
			}
			targets[i] = t
		default:
			return nil, fmt.Errorf("unknown target type: %s", target.Type)
		}
	}
	return &Backup{
		cfg:     cfg,
		sources: sources,
		targets: targets,
	}, nil
}

func (b *Backup) Run() error {
	backupDest, err := os.MkdirTemp("", "backup")
	if err != nil {
		return fmt.Errorf("failed to create backup destination: %w", err)
	}

	encryptedFileName := fmt.Sprintf("backup_%s.bin", time.Now().Format("2006-01-02_15-04-05"))
	encryptedFileDest := filepath.Join(backupDest, encryptedFileName)
	err = b.Encrypt(encryptedFileDest)
	if err != nil {
		return fmt.Errorf("failed to encrypt backup: %w", err)
	}

	for i, target := range b.targets {
		err := target.Upload(encryptedFileDest)
		if err != nil {
			return fmt.Errorf("failed to upload backup: %w", err)
		}
		err = target.Clean(b.cfg.Targets[i].BackupExpirationDays)
		if err != nil {
			return fmt.Errorf("failed to clean old backups in target: %w", err)
		}
	}

	return nil
}

func (b *Backup) Encrypt(encryptedFilePath string) error {
	backupDest, err := os.MkdirTemp("", "backup")
	if err != nil {
		return fmt.Errorf("failed to create backup destination: %w", err)
	}
	for _, source := range b.sources {
		err := source.Backup(backupDest)
		if err != nil {
			return fmt.Errorf("failed to backup source: %w", err)
		}
	}
	log.Debug().Str("destination", backupDest).Msg("backuped up data")

	compressedBackupFileName := fmt.Sprintf("backup_%s.zip", time.Now().Format("2006-01-02_15-04-05"))
	compressedBackupDest := filepath.Join(backupDest, compressedBackupFileName)
	err = compress(backupDest, compressedBackupDest)
	if err != nil {
		return fmt.Errorf("failed to compress backup: %w", err)
	}
	log.Debug().Str("destination", compressedBackupDest).Msg("compressed backup")

	err = delete(backupDest, compressedBackupFileName)
	if err != nil {
		return fmt.Errorf("failed to delete uncompressed backup: %w", err)
	}
	log.Debug().Msg("deleted uncompressed backup")

	err = encrypt(compressedBackupDest, encryptedFilePath, b.cfg.EncryptionPassword)
	if err != nil {
		return fmt.Errorf("failed to encrypt backup: %w", err)
	}
	log.Debug().Str("destination", encryptedFilePath).Msg("encrypted backup")
	return nil
}

func compress(folder, destination string) error {
	// create a new zip archive
	zipFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	// walk the folder and add files to the zip archive
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == folder || path == destination {
			return nil
		}

		// create a new file header for the file
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set the file name to be relative to the source directory
		header.Name = strings.TrimPrefix(path, folder+string(os.PathSeparator))

		if info.IsDir() {
			// if the file is a directory, create it in the archive with an empty header
			header.Name += "/"
			_, err = archive.CreateHeader(header)
			if err != nil {
				return err
			}
		} else {
			// if the file is not a directory, create it in the archive with its contents
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func encrypt(file string, encryptedFile string, password string) error {
	// read the plaintext file
	fileData, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// generate a 32-byte key and a 16-byte initialization vector from the password
	key, iv := generateKeyAndIV(password)

	// create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// create a new CBC mode block cipher with the AES cipher and the initialization vector
	mode := cipher.NewCBCEncrypter(block, iv)

	// pad the plaintext to a multiple of the block size
	fileData = pad(fileData, block.BlockSize())

	// create a new buffer to hold the encrypted data
	ciphertext := make([]byte, len(fileData))

	// encrypt the plaintext using the CBC mode block cipher
	mode.CryptBlocks(ciphertext, fileData)

	// write the encrypted data to the encrypted file
	err = os.WriteFile(encryptedFile, ciphertext, 0644)
	if err != nil {
		return err
	}

	return nil
}

func Decrypt(encryptedFile string, decryptedFile string, password string) error {
	// read the encrypted file
	ciphertext, err := os.ReadFile(encryptedFile)
	if err != nil {
		return err
	}

	// generate a 32-byte key and a 16-byte initialization vector from the password
	key, iv := generateKeyAndIV(password)

	// create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// create a new CBC mode block cipher with the AES cipher and the initialization vector
	mode := cipher.NewCBCDecrypter(block, iv)

	// create a new buffer to hold the decrypted data
	fileData := make([]byte, len(ciphertext))

	// decrypt the ciphertext using the CBC mode block cipher
	mode.CryptBlocks(fileData, ciphertext)

	// unpad the decrypted data
	fileData = unpad(fileData)

	// write the decrypted data to the decrypted file
	err = os.WriteFile(decryptedFile, fileData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func Restore(backupFile string, restoreDest string, password string) error {
	// decrypt the backup file
	decryptedFilePath := filepath.Dir(backupFile)
	decryptedFilePath = filepath.Join(decryptedFilePath, strings.TrimSuffix(filepath.Base(backupFile), ".bin")+".zip")
	err := Decrypt(backupFile, decryptedFilePath, password)
	if err != nil {
		return fmt.Errorf("failed to decrypt backup: %w", err)
	}

	// decompress the backup file
	err = Decompress(decryptedFilePath, restoreDest)
	if err != nil {
		return fmt.Errorf("failed to decompress backup: %w", err)
	}

	// delete the decrypted backup file
	err = os.Remove(decryptedFilePath)
	if err != nil {
		return fmt.Errorf("failed to delete decrypted backup: %w", err)
	}

	return nil
}

func Decompress(backupFile string, destination string) error {
	// open the zip archive
	zipReader, err := zip.OpenReader(backupFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// walk the files in the archive
	for _, file := range zipReader.File {
		// create a new file in the destination
		filePath := filepath.Join(destination, file.Name)
		if file.FileInfo().IsDir() {
			// if the file is a directory, create it
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			// if the file is not a directory, create it and copy the contents
			fileReader, err := file.Open()
			if err != nil {
				return err
			}
			defer fileReader.Close()

			fileWriter, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer fileWriter.Close()

			_, err = io.Copy(fileWriter, fileReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// unpad removes the PKCS7 padding from the input
func unpad(input []byte) []byte {
	paddingSize := int(input[len(input)-1])
	return input[:len(input)-paddingSize]
}

// generateKeyAndIV generates a 32-byte key and a 16-byte initialization vector from a password
func generateKeyAndIV(password string) ([]byte, []byte) {
	key := make([]byte, 32)
	iv := make([]byte, 16)

	// derive the key and the initialization vector from the password using PBKDF2
	iterations := 10000 // adjust this to balance security and performance
	keyBytes := pbkdf2.Key([]byte(password), []byte{}, iterations, 32, sha256.New)
	ivBytes := pbkdf2.Key([]byte(password), []byte{}, iterations, 16, sha256.New)

	copy(key, keyBytes)
	copy(iv, ivBytes)

	return key, iv
}

// pad pads the input to a multiple of the block size using PKCS7 padding
func pad(input []byte, blockSize int) []byte {
	paddingSize := blockSize - len(input)%blockSize
	padding := bytes.Repeat([]byte{byte(paddingSize)}, paddingSize)
	return append(input, padding...)
}

func delete(folderPath, whitelistedFileName string) error {
	// get a list of all files and folders inside the specified folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// loop through each file and folder
	for _, file := range files {
		// check if the current file is the one to keep
		if file.Name() == whitelistedFileName {
			continue
		}

		// create the full path to the current file
		filePath := filepath.Join(folderPath, file.Name())

		// check if the current file is a directory
		if file.IsDir() {
			// if it is, delete the directory and its contents
			err := os.RemoveAll(filePath)
			if err != nil {
				return fmt.Errorf("failed to remove directory: %w", err)
			}
		} else {
			// if it's a file, delete the file
			err := os.Remove(filePath)
			if err != nil {
				fmt.Println(err)
				return fmt.Errorf("failed to remove file: %w", err)
			}
		}
	}

	return nil
}
