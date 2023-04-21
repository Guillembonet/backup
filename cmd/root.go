package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup is a tool to encrypt and backup your files",
}

func init() {
	rootCmd.AddCommand(decryptCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(backupCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
