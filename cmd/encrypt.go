package cmd

import (
	"fmt"

	"github.com/guillembonet/backup/backup"
	"github.com/guillembonet/backup/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt your files",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, err := cmd.Flags().GetString("config-path")
		if err != nil {
			log.Fatal().Err(err).Msg("no config path defined")
		}
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load config")
		}

		if cfg.Runtime.LogLevel == "" {
			cfg.Runtime.LogLevel = "debug"
		}
		logLevel, err := zerolog.ParseLevel(cfg.Runtime.LogLevel)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse log level")
		}
		log.Logger = log.Logger.Level(logLevel)

		log.Info().Str("config", fmt.Sprintf("%+v", cfg)).Msg("loaded config")

		backup, err := backup.New(cfg.Backup)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create backup service")
		}

		outputPath, err := cmd.Flags().GetString("output-path")
		if err != nil {
			log.Fatal().Err(err).Msg("no output path defined")
		}

		err = backup.Encrypt(outputPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to encrypt files")
		}

		log.Info().
			Str("output_path", outputPath).
			Msg("successfully encrypted files")
	},
}

func init() {
	encryptCmd.Flags().StringP("config-path", "c", "./example_config.yaml", "config file path")
	encryptCmd.Flags().StringP("output-path", "o", "./backup.bin", "output file path of resulting encrypted file")
}
