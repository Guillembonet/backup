package cmd

import (
	"github.com/guillembonet/backup/backup"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var decryptCmd = &cobra.Command{
	Use:  "decrypt [encrypted file path]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputFile, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Fatal().Err(err).Msg("no output file defined")
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			log.Fatal().Err(err).Msg("no password defined")
		}

		encryptedFilePath := args[0]
		err = backup.Decrypt(encryptedFilePath, outputFile, password)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to decrypt")
		}
		log.Info().Msg("decrypted")
	},
}

func init() {
	decryptCmd.Flags().StringP("output", "o", "./decrypted_file.zip", "output file path")
	decryptCmd.Flags().StringP("password", "p", "", "password for encryption/decryption")
}
