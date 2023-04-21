package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/guillembonet/backup/backup"
	"github.com/guillembonet/backup/config"
)

func main() {
	cfg, err := config.LoadConfig()
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

	err = backup.Run()
	if err != nil {
		if cfg.RunMode.RunOnceAndExit {
			log.Fatal().Err(err).Msg("failed to run backup service")
		}
		log.Error().Err(err).Msg("failed to run backup service")
	}
	if cfg.RunMode.RunOnceAndExit {
		os.Exit(0)
	}

	sigKill := make(chan os.Signal, 1)
	signal.Notify(sigKill, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-sigKill:
			log.Info().Msg("received kill signal, exiting")
			os.Exit(0)
		case <-time.After(cfg.RunMode.Interval):
			err = backup.Run()
			if err != nil {
				log.Error().Err(err).Msg("failed to run backup service")
			}
		}
	}
}
