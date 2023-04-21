package mega

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	mega "github.com/t3rm1n4l/go-mega"
)

type Client struct {
	client *mega.Mega
	cfg    map[string]string
}

func NewClient(cfg map[string]string) (*Client, error) {
	return &Client{
		client: mega.New(),
		cfg:    cfg,
	}, nil
}

func (c *Client) Upload(filePath string) error {
	username, ok := c.cfg["username"]
	if !ok {
		return fmt.Errorf("missing username")
	}
	password, ok := c.cfg["password"]
	if !ok {
		return fmt.Errorf("missing password")
	}
	backupFolder, ok := c.cfg["backup_folder"]
	if !ok {
		return fmt.Errorf("missing backup_folder")
	}

	err := c.client.Login(username, password)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	log.Debug().Msg("logged in")

	backupsNode, err := c.getBackupNode(backupFolder)
	if err != nil {
		return fmt.Errorf("failed to create backups dir: %w", err)
	}
	log.Debug().Msg("created backups dir")

	fileNode, err := c.client.UploadFile(filePath, backupsNode, filepath.Base(filePath), nil)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.Debug().Str("name", fileNode.GetName()).
		Str("timestamp", fileNode.GetTimeStamp().String()).
		Int64("size", fileNode.GetSize()).
		Msg("uploaded file")

	return nil
}

func (c *Client) Clean(backupExpirationDays int) error {
	username, ok := c.cfg["username"]
	if !ok {
		return fmt.Errorf("missing username")
	}
	password, ok := c.cfg["password"]
	if !ok {
		return fmt.Errorf("missing password")
	}
	backupFolder, ok := c.cfg["backup_folder"]
	if !ok {
		return fmt.Errorf("missing backup_folder")
	}

	err := c.client.Login(username, password)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	log.Debug().Msg("logged in")

	backupsNode, err := c.getBackupNode(backupFolder)
	if err != nil {
		return fmt.Errorf("failed to create backups dir: %w", err)
	}

	children, err := c.client.FS.GetChildren(backupsNode)
	if err != nil {
		return fmt.Errorf("failed to get children: %w", err)
	}

	for _, child := range children {
		if child.GetTimeStamp().Before(time.Now().AddDate(0, 0, -backupExpirationDays)) {
			err = c.client.Delete(child, false)
			if err != nil {
				return fmt.Errorf("failed to delete file: %w", err)
			}
			log.Debug().Str("name", child.GetName()).
				Str("timestamp", child.GetTimeStamp().String()).
				Int64("size", child.GetSize()).
				Msg("deleted old backup")
		}
	}
	return nil
}

func (c *Client) getBackupNode(folderName string) (*mega.Node, error) {
	root := c.client.FS.GetRoot()
	rootChildern, err := c.client.FS.GetChildren(root)
	if err != nil {
		return nil, fmt.Errorf("failed to get root's children: %w", err)
	}
	for _, child := range rootChildern {
		if child.GetName() == folderName {
			return child, nil
		}
	}

	//create if not found
	backupsNode, err := c.client.CreateDir(folderName, root)
	if err != nil {
		return nil, fmt.Errorf("failed to create backups dir: %w", err)
	}

	return backupsNode, nil
}
