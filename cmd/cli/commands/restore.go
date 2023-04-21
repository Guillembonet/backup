package commands

import (
	"github.com/guillembonet/backup/backup"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [encrypted file path]",
	Short: "restore a backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputDir, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Fatal().Err(err).Msg("no output directory defined")
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			log.Fatal().Err(err).Msg("no password defined")
		}

		encryptedFilePath := args[0]
		err = backup.Restore(encryptedFilePath, outputDir, password)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to restore")
		}
		log.Info().Msg("restored")
	},
}

func init() {
	restoreCmd.Flags().StringP("output", "o", "./", "output directory")
	restoreCmd.Flags().StringP("password", "p", "", "password for encryption/decryption")
}
